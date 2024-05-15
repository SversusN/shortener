package app

import (
	"fmt"
	"go.uber.org/zap"

	"log"
	"net/http"

	"github.com/SversusN/shortener/config"
	"github.com/SversusN/shortener/internal/handlers"
	"github.com/SversusN/shortener/internal/logger"
	"github.com/SversusN/shortener/internal/storage/primitivestorage"
	"github.com/SversusN/shortener/internal/storage/storage"
	"github.com/go-chi/chi/v5"
)

type App struct {
	Config   *config.Config
	Storage  storage.Storage
	Handlers *handlers.Handlers
	Logger   *logger.ServerLogger
}

func New() *App {
	cfg := config.NewConfig()
	ns := primitivestorage.NewStorage()
	nh := handlers.NewHandlers(cfg, ns)
	lg := logger.CreateLogger(zap.NewAtomicLevelAt(zap.InfoLevel)) //Хардкод TODO
	return &App{cfg, ns, nh, lg}
}

func (a App) CreateRouter(hnd handlers.Handlers) chi.Router {
	r := chi.NewRouter()
	//r.Use(mw.New(r))
	r.Use(a.Logger.LoggingMW())
	r.Route("/", func(r chi.Router) {
		r.Post("/", hnd.HandlerPost)
		r.Get("/{shortKey}", hnd.HandlerGet)
	})
	return r
}

func (a App) Run() {
	r := a.CreateRouter(*a.Handlers)
	//go client.GetClient("http://" + a.Config.FlagAddress)
	fmt.Printf("running on %s\n", a.Config.FlagAddress)
	log.Fatal(
		http.ListenAndServe(a.Config.FlagAddress, r), "упали...")
}
