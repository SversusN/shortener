package app

import (
	"github.com/SversusN/shortener/config"
	"github.com/SversusN/shortener/internal/handlers"
	"github.com/SversusN/shortener/internal/storage/primitivestorage"
	"github.com/SversusN/shortener/internal/storage/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type App struct {
	Config   *config.Config
	Storage  storage.Storage
	Handlers *handlers.Handlers
	Router   chi.Router
}

func New() *App {
	appconfig := config.NewConfig()
	newStorage := primitivestorage.NewStorage()
	apphandlers := handlers.NewHandlers(appconfig, newStorage)

	return &App{appconfig,
		newStorage,
		apphandlers,
		CreateRouter(apphandlers.HandlerGet, apphandlers.HandlerPost),
	}
}

func CreateRouter(g http.HandlerFunc, p http.HandlerFunc) chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", p)
		r.Get("/{shortKey}", g)
	})
	return r
}
