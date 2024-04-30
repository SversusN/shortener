package config

import (
	"errors"
	"flag"
	"strings"
)

type Config struct {
	FlagAddress     string
	FlagBaseAddress string
}

func NewConfig() (c *Config) {
	c = &Config{
		FlagAddress:     ":8080",
		FlagBaseAddress: "http://localhost:8080",
	}

	// указываем ссылку на переменную, имя флага, значение по умолчанию и описание
	flag.StringVar(&c.FlagAddress, "a", c.FlagAddress, "Устанавливаем ip адрес нашего сервера.")
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
	flag.Parse()
	//if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
	//	c.FlagAddress = envRunAddr
	//}
	//if envRunAddr := os.Getenv("BASE_URL"); envRunAddr != "" {
	//	c.FlagBaseAddress = envRunAddr
	//}
	return c
}
