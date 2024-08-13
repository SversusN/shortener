package app

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/SversusN/shortener/config"
	"github.com/SversusN/shortener/internal/handlers"
	"github.com/SversusN/shortener/internal/logger"
	mw "github.com/SversusN/shortener/internal/middleware"
	"github.com/SversusN/shortener/internal/pkg/utils"
	"github.com/SversusN/shortener/internal/storage/dbstorage"
	"github.com/SversusN/shortener/internal/storage/primitivestorage"
	"github.com/SversusN/shortener/internal/storage/storage"
)

type App struct {
	Config     *config.Config
	Storage    storage.Storage
	Handlers   *handlers.Handlers
	Logger     *logger.ServerLogger
	FileHelper *utils.FileHelper
	Context    context.Context
}

func New() *App {
	var ns storage.Storage
	cfg := config.NewConfig()
	ctx := context.Background() //TODO https://habr.com/ru/articles/771626/
	fh, err := utils.NewFileHelper(cfg.FlagFilePath)
	//DB
	if cfg.DataBaseDSN == "" {
		ns = primitivestorage.NewStorage(fh, err)
	} else {
		ns, err = dbstorage.NewDB(ctx, cfg.DataBaseDSN)
		if err != nil {
			log.Fatalln("Failed to connect to database", err)
		}
	}
	nh := handlers.NewHandlers(cfg, ns)
	lg := logger.CreateLogger(zap.NewAtomicLevelAt(zap.InfoLevel))

	return &App{cfg, ns, nh, lg, fh, ctx}
}

func (a App) CreateRouter(hnd handlers.Handlers) chi.Router {
	r := chi.NewRouter()
	r.Use(a.Logger.LoggingMW())
	r.Use(mw.GzipMiddleware)
	r.Use(mw.NewAuthMW().AuthMWfunc)
	r.Route("/", func(r chi.Router) {
		r.Post("/", hnd.HandlerPost)
		r.Get("/ping", hnd.HandlerDBPing)
		r.Get("/{shortKey}", hnd.HandlerGet)
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", hnd.HandlerJSONPost)
			r.Post("/shorten/batch", hnd.HandlerJSONPostBatch)
			r.Group(func(r chi.Router) { //secure
				r.Get("/user/urls", hnd.HandlerGetUserURLs)
				r.Delete("/user/urls", hnd.HandlerDeleteUserURLs)
			})

		})
	})
	return r
}

func (a App) Run() {
	r := a.CreateRouter(*a.Handlers)
	go func() {
		log.Println(http.ListenAndServe("localhost:90", nil))
	}()
	log.Printf("running on %s\n", a.Config.FlagAddress)
	log.Fatal(
		http.ListenAndServe(a.Config.FlagAddress, r), "упали...")
}
