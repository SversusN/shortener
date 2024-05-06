package main

import (
	_ "github.com/SversusN/shortener/docs"
	"github.com/SversusN/shortener/internal/app"
)

// @title Swagger shortener API
// @version 1.0
// @description This is a sample shortener server.
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @Host: "localhost:80"
// @BasePath /
func main() {
	app.New().Run()
}
