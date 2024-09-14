// App собирает зависимости приложения (конфигурация, хранилище, логгер)
package app

import (
	"context"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

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

// App структура приложения
type App struct {
	Config     *config.Config       // Объект конфигурации
	Storage    storage.Storage      // Интерфейс хранилища
	Handlers   *handlers.Handlers   //Объект http обработчиков
	Logger     *logger.ServerLogger //Внедорение логера
	FileHelper *utils.FileHelper    //Работа с файлом
	Context    context.Context      //Контекст приложения
}

// App Конструктор пакета, создает целевой объект приложения с нужными зависимостями
func New() *App {
	var ns storage.Storage
	cfg := config.NewConfig()
	ctx := context.Background()
	fh, err := utils.NewFileHelper(cfg.FlagFilePath)
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

// CreateRouter Создание роутера Chi
func (a App) CreateRouter(hnd handlers.Handlers) chi.Router {
	r := chi.NewRouter()
	r.Use(a.Logger.LoggingMW())
	r.Use(mw.GzipMiddleware)
	r.Use(mw.NewAuthMW().AuthMWfunc)
	//Инициализация маршрута для роутера Chi
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

// Run Создание роутера веб сервера и запуск веб сервера
func (a App) Run() {
	r := a.CreateRouter(*a.Handlers)
	//Переменные для завершения
	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		log.Println(http.ListenAndServe("localhost:90", nil))
	}()

	log.Printf("running on %s\n", a.Config.FlagAddress)
	server := &http.Server{
		Addr:    a.Config.FlagAddress,
		Handler: r,
	}
	//Ждем сигнала завершения
	go func() {
		<-sigint
		log.Println("shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Println("HTTP server Shutdown:", err)
		}
		close(idleConnsClosed)
	}()

	if a.Config.EnableHTTPS {
		manager := &autocert.Manager{
			// директория для хранения сертификатов
			Cache: autocert.DirCache("cache-dir"),
			// функция, принимающая Terms of Service издателя сертификатов
			Prompt: autocert.AcceptTOS,
			// перечень доменов, для которых будут поддерживаться сертификаты
			HostPolicy: autocert.HostWhitelist("example.com"),
		}
		// конструируем сервер с поддержкой TLS
		server = &http.Server{
			Addr:    ":443",
			Handler: r,
			// для TLS-конфигурации используем менеджер сертификатов
			TLSConfig: manager.TLSConfig(),
		}
		<-idleConnsClosed
		log.Println("Server Shutdown gracefully")

		//запуск https
		if err := server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			// ошибки старта или остановки Listener
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
		<-idleConnsClosed
		log.Println("Server Shutdown gracefully")
	} else {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			// ошибки старта или остановки Listener
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}
	<-idleConnsClosed
	log.Println("Server Shutdown gracefully")

}
