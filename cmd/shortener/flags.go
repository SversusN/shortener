package main

import (
	"errors"
	"flag"
	"strings"
)

var (
	flagAddress     string
	flagBaseAddress string
)

func ParseFlags() {

	flagBaseAddress = "http://localhost:8080"

	flag.StringVar(&flagAddress, "a", "localhost:8080", "Address:port hosting shortener service")
	//так как в тесте бывает http нужно распарсить значение
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
}
