// Package grpcsrv содержит реализацию gRPC сервера.
package grpcsrv

import (
	"context"
	"errors"
	"google.golang.org/grpc/metadata"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/SversusN/shortener/config"
	"github.com/SversusN/shortener/internal/grpcsrv/interceptors"
	pb "github.com/SversusN/shortener/internal/grpcsrv/proto"
	"github.com/SversusN/shortener/internal/internalerrors"
	"github.com/SversusN/shortener/internal/pkg/utils"
	entity "github.com/SversusN/shortener/internal/storage/dbstorage"
	"github.com/SversusN/shortener/internal/storage/storage"
)

// ShortenerServer описывает тип gRPC сервера.
type ShortenerServer struct {
	pb.UnimplementedShortenerServer
	ctx     *context.Context
	storage storage.Storage
	cfg     *config.Config
	wg      *sync.WaitGroup
}

// NewGRPCServer создает и возвращает новый сервер.
func NewGRPCServer(ctx *context.Context, storage storage.Storage, cfg *config.Config, wg *sync.WaitGroup) *grpc.Server {
	authInterceptor := interceptors.NewAuthInterceptor(*ctx)
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.LoggerInterceptor, authInterceptor.AuthenticateUser),
	)
	pb.RegisterShortenerServer(s, &ShortenerServer{ctx: ctx, storage: storage, cfg: cfg, wg: wg})
	return s
}

// ShortenURL обрабатывает запрос на сокращение ссылки.
func (s *ShortenerServer) ShortenURL(ctx context.Context, in *pb.URLRequest) (*pb.URLResponse, error) {
	var response pb.URLResponse
	userID, err := utils.GetUserIDFromCtx(ctx, "user_id")
	if errors.Is(err, internalerrors.ErrUserTypeError) {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	var shortURL string
	key := utils.GenerateShortKey()

	shortURL, err = s.storage.SetURL(key, in.GetOriginalUrl(), userID)
	if err != nil {
		switch {
		case errors.Is(err, internalerrors.ErrKeyAlreadyExists):
			return nil, status.Error(codes.InvalidArgument, "Некорректная ключ для сокращения")
		case errors.Is(err, internalerrors.ErrOriginalURLAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "Ссылыка уже была сохранена")
		default:
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	}
	response.ShortUrl = shortURL
	return &response, nil
}

// ShortenBatchURL обрабатывает пакетный запрос на сокращение ссылок.
func (s *ShortenerServer) ShortenBatchURL(ctx context.Context, in *pb.BatchURLRequest) (*pb.BatchURLResponse, error) {

	userID, err := utils.GetUserIDFromCtx(ctx, "user_id")
	if errors.Is(err, internalerrors.ErrUserTypeError) {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	saveUrls := make(map[string]entity.UserURL)
	for _, url := range in.GetUrls() {
		newkey := utils.GenerateShortKey()
		saveUrls[newkey] = entity.UserURL{UserID: userID, OriginalURL: url.OriginalUrl}
	}
	savedBatch, err := s.storage.SetURLBatch(saveUrls)

	if err != nil {
		switch {
		case errors.Is(err, internalerrors.ErrKeyAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "В запросе ссылки, которые уже были ранее сохранены")
		default:
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	}
	response := pb.BatchURLResponse{Urls: []*pb.BatchURLResponse_BatchURL{}}
	for u := range savedBatch {
		response.Urls = append(response.Urls, &pb.BatchURLResponse_BatchURL{
			ShortUrl:      utils.GetFullURL(s.cfg.FlagBaseAddress, u),
			CorrelationId: u,
		})
	}
	return &response, nil
}

// GetURL обрабатывает запрос на получение полной ссылки по сокращенному id.
func (s *ShortenerServer) GetURL(ctx context.Context, in *pb.GetURLReq) (*pb.GetURLRes, error) {
	var response pb.GetURLRes
	url, err := s.storage.GetURL(in.GetUrlId())
	if err != nil {
		switch {
		case errors.Is(err, internalerrors.ErrDeleted):
			return nil, status.Error(codes.NotFound, "Ссылка удалена")
		case errors.Is(err, internalerrors.ErrNotFound):
			return nil, status.Error(codes.NotFound, "Ссылка не найдена")
		default:
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	}
	response.OriginalUrl = url
	return &response, nil
}

// GetUserURLs обрабатывает запрос на получение ссылок, сокращенных пользователем.
func (s *ShortenerServer) GetUserURLs(ctx context.Context, _ *pb.GetUsersURLsReq) (*pb.GetUsersURLsRes, error) {
	response := pb.GetUsersURLsRes{
		Urls: []*pb.GetUsersURLsRes_UserURL{},
	}
	userID, err := utils.GetUserIDFromCtx(ctx, "user_id")
	if errors.Is(err, internalerrors.ErrUserTypeError) {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	userURLs, err := s.storage.GetUserUrls(userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	entities, ok := userURLs.([]entity.UserURLEntity)
	if !ok {
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	for _, url := range entities {
		response.Urls = append(response.Urls, &pb.GetUsersURLsRes_UserURL{
			OriginalUrl: url.OriginalURL,
			ShortUrl:    url.ShortURL,
		})
	}
	return &response, nil
}

// DeleteUserURLs обрабатывает запрос на удаление ссылок пользователя.
func (s *ShortenerServer) DeleteUserURLs(ctx context.Context, in *pb.DeleteUserURLsReq) (*pb.DeleteUserURLsRes, error) {
	var response pb.DeleteUserURLsRes
	userID, err := utils.GetUserIDFromCtx(ctx, "user_id")
	if errors.Is(err, internalerrors.ErrUserTypeError) {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	deleteCh, err := s.storage.DeleteUserURLs(userID, s.wg)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer close(deleteCh)
		for _, key := range in.GetUrls() {
			deleteCh <- key
		}
	}()
	if err != nil {
		if errors.Is(err, internalerrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "no data to delete")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &response, nil
}

// GetStats обрабатывает запрос на получение статистики хранилища.
func (s *ShortenerServer) GetStats(ctx context.Context, _ *pb.GetStatsReq) (*pb.GetStatsRes, error) {
	ts, err := utils.GetCIDR(s.cfg.TrustedSubnet)
	if err != nil {
		ts = nil
	}
	if ts == nil {
		return nil, status.Error(codes.PermissionDenied, "Forbidden")
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("X-Real-IP")
		if len(values) > 0 {
			ip := net.ParseIP(values[0])
			if ip != nil && !ts.Contains(ip) {
				return nil, status.Error(codes.PermissionDenied, "Forbidden")
			}
		}
	}
	var response pb.GetStatsRes
	urls, users, err := s.storage.GetStats()
	if err != nil {
		response.Urls = int32(urls)
		response.Users = int32(users)
		return &response, nil
	} else {
		return nil, status.Error(codes.Internal, "forbidden")
	}
}

// Ping обрабатывает запрос на проверку соединения с хранилищем данных.
func (s *ShortenerServer) Ping(ctx context.Context, _ *pb.PingRequest) (*pb.PingResponse, error) {
	var response pb.PingResponse
	if err := s.storage.(storage.Pinger).Ping(); err != nil {
		return nil, status.Error(codes.Internal, "Failed to ping database")
	}
	return &response, nil
}
