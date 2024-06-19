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
	DataBaseDSN     string
}

func NewConfig() (c *Config) {
	currentDir, _ := os.Getwd() //плоховато работает на винде
	c = &Config{
		FlagAddress:     ":8080",
		FlagBaseAddress: "http://localhost:8080",
		FlagFilePath:    fmt.Sprint(currentDir, "/tmp/short-url-db.json"),
		DataBaseDSN:     os.Getenv("DATABASE_DSN"),
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
			c.FlagBaseAddress = fmt.Sprint("http://", flagValue)
		}
		return nil
	})
	flag.StringVar(&c.FlagFilePath, "f", c.FlagFilePath, "set file path")
	flag.StringVar(&c.DataBaseDSN, "d", c.DataBaseDSN, "Database connection string")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		c.FlagAddress = envRunAddr
	}
	if envBaseAddr := os.Getenv("BASE_URL"); envBaseAddr != "" {
		c.FlagBaseAddress = envBaseAddr
	}
	//может быть ""
	if envFilePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		c.FlagFilePath = envFilePath
	}
	if dataBaseDSN, ok := os.LookupEnv("DATABASE_DSN"); ok {
		c.DataBaseDSN = dataBaseDSN
	}
	return c
}
