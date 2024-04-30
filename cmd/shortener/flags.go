package main

import (
	"errors"
	"flag"
	"os"
	"strings"
)

var (
	flagAddress     string
	flagBaseAddress = "http://localhost:8080"
)

func ParseFlags() {
	flag.StringVar(&flagAddress, "a", ":8080", "Address:port hosting shortener service")
	//так как в тесте может быть http(возможно s) нужно распарсить значение
	flag.Func("b", "URL for redirect", func(flagValue string) error {
		hp := strings.Split(flagValue, ":")
		if len(hp) < 2 {
			return errors.New("need address in a form host:port")
		}
		if hp[0] == "http" || hp[0] == "https" {
			flagBaseAddress = flagValue
		} else {
			flagBaseAddress = "http://" + flagValue
		}

		return nil
	})
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		flagAddress = envRunAddr
	}
	if envRunAddr := os.Getenv("BASE_URL"); envRunAddr != "" {
		flagBaseAddress = envRunAddr
	}

}
