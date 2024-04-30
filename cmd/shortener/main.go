package main

import (
	"fmt"
	"github.com/SversusN/shortener/internal/app"
	"log"
	"net/http"
)

func main() {
	a := app.New()
	fmt.Printf("running on %s\n", a.Config.FlagAddress)
	log.Fatal(
		http.ListenAndServe(a.Config.FlagAddress, a.Router), "упали...")
}
