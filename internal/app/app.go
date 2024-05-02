package app

import (
	"fmt"
	"github.com/SversusN/shortener/config"
	"github.com/SversusN/shortener/internal/handlers"
	mw "github.com/SversusN/shortener/internal/middleware"
	"github.com/SversusN/shortener/internal/storage/primitivestorage"
	"github.com/SversusN/shortener/internal/storage/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type App struct {
	Config   *config.Config
	Storage  storage.Storage
	Handlers *handlers.Handlers
}

func New() *App {
	cfg := config.NewConfig()
	ns := primitivestorage.NewStorage()
	nh := handlers.NewHandlers(cfg, ns)
	return &App{cfg, ns, nh}
}

func (a App) CreateRouter(hnd handlers.Handlers) chi.Router {
	r := chi.NewRouter()
	r.Use(mw.New(r))
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
