package app

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/SversusN/shortener/config"
	"github.com/SversusN/shortener/internal/handlers"
	"github.com/SversusN/shortener/internal/logger"
	mw "github.com/SversusN/shortener/internal/middleware"
	"github.com/SversusN/shortener/internal/pkg/utils"
	"github.com/SversusN/shortener/internal/storage/primitivestorage"
	"github.com/SversusN/shortener/internal/storage/storage"
)

type App struct {
	Config     *config.Config
	Storage    storage.Storage
	Handlers   *handlers.Handlers
	Logger     *logger.ServerLogger
	FileHelper *utils.FileHelper
}

func New() *App {
	cfg := config.NewConfig()
	fh, err := utils.NewFileHelper(cfg.FlagFilePath)
	ns := primitivestorage.NewStorage(fh, err)
	nh := handlers.NewHandlers(cfg, ns)
	lg := logger.CreateLogger(zap.NewAtomicLevelAt(zap.InfoLevel)) //Хардкод TODO
	return &App{cfg, ns, nh, lg, fh}
}

func (a App) CreateRouter(hnd handlers.Handlers) chi.Router {
	r := chi.NewRouter()
	r.Use(a.Logger.LoggingMW())
	r.Use(mw.GzipMiddleware)
	r.Route("/", func(r chi.Router) {
		r.Post("/", hnd.HandlerPost)
		r.Get("/{shortKey}", hnd.HandlerGet)
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", hnd.HandlerJSONPost)
		})
	})
	return r
}

func (a App) Run() {
	r := a.CreateRouter(*a.Handlers)
	log.Printf("running on %s\n", a.Config.FlagAddress)
	log.Fatal(
		http.ListenAndServe(a.Config.FlagAddress, r), "упали...")
}
