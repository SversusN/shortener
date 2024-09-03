// Пакет main создает и запускает приложение.
package main

import (
	"fmt"

	"github.com/SversusN/shortener/internal/app"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Run version: %s,\nBuild date: %s,\nBuild commit: %s,\n",
		buildVersion,
		buildDate,
		buildCommit,
	)
	app.New().Run()
}
