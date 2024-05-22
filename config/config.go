package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	FlagAddress     string
	FlagBaseAddress string
	FlagFilePath    string
}

func NewConfig() (c *Config) {
	currentDir, _ := os.Getwd() //плоховато работает на винде
	c = &Config{
		FlagAddress:     ":8080",
		FlagBaseAddress: "http://localhost:8080",
		FlagFilePath:    fmt.Sprint(currentDir, "/tmp/short-url-db.json"),
	}
	// указываем ссылку на переменную, имя флага, значение по умолчанию и описание
	flag.StringVar(&c.FlagAddress, "a", c.FlagAddress, "set server IP address")
	flag.Func("b", "URL for redirect", func(flagValue string) error {
		hp := strings.Split(flagValue, ":")
		if len(hp) < 2 {
			return errors.New("need address in a form host:port")
		}
		if hp[0] == "http" || hp[0] == "https" {
			c.FlagBaseAddress = flagValue
		} else {
			c.FlagBaseAddress = "http://" + flagValue
		}
		return nil
	})
	flag.StringVar(&c.FlagFilePath, "f", c.FlagFilePath, "set file path")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		c.FlagAddress = envRunAddr
	}
	if envBaseAddr := os.Getenv("BASE_URL"); envBaseAddr != "" {
		c.FlagBaseAddress = envBaseAddr
	}
	//if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
	//	c.FlagFilePath = envFilePath
	//} else {
	//	c.FlagAddress = envFilePath
	//}
	//может быть ""
	c.FlagFilePath = os.Getenv("FILE_STORAGE_PATH")

	return c
}
