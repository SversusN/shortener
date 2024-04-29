package main

import (
	"flag"
)

var flagAdress string
var flagBaseAdress string

func ParseFlags() {
	flag.StringVar(&flagAdress, "a", "localhost:8080", "Adress:port hosting shortener service")
	flag.StringVar(&flagBaseAdress, "b", "localhost:8080", "Adress:port for redirict")
	flag.Parse()
}
