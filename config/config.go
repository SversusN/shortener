// Config содержит поля конфигурации.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// Cfg, Config для работы с JSON файлом
var (
	Cfg        Config // Cfg - переменная с конфигурацией
	ConfigPath string
)

// Config структура с полями конфигурации
type Config struct {
	FlagAddress     string `json:"flag_address"`      //Адрес запуска сервера
	FlagBaseAddress string `json:"flag_base_address"` //Базовый URL
	FlagFilePath    string `json:"flag_file_path"`    //Флаг для хранения файла
	DataBaseDSN     string `json:"data_base_dsn"`     //Строка соединения с БД
	EnableHTTPS     bool   `json:"enable_https"`      //Подключение tls соединения
	TrustedSubnet   string `json:"trusted_subnet"`    //доверенная посеть
	GRPCAddress     string `json:"grpc_address"`      //Адрес сервера grpc
}

// NewConfig конструктор для внедрения зависимостей
func NewConfig() (c *Config) {
	currentDir, _ := os.Getwd() //плоховато работает на винде
	//Инициализация с переменными по умолчанию
	c = &Config{
		FlagAddress:     ":8080",
		FlagBaseAddress: "http://localhost:8080",
		FlagFilePath:    fmt.Sprint(currentDir, "/tmp/short-url-db.json"),
		DataBaseDSN:     os.Getenv("DATABASE_DSN"),
		EnableHTTPS:     false,
		TrustedSubnet:   "",
		GRPCAddress:     "3200",
	}
	// Получаем путь к конфигурационному файлу из переменных окружения, если указан
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		ConfigPath = configPath
		err := loadConfigFromFile()
		if err != nil {
			return
		}
	}
	// Указываем ссылку на переменную, имя флага, значение по умолчанию и описание
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
	flag.BoolVar(&c.EnableHTTPS, "s", c.EnableHTTPS, "Enable secure connection")
	flag.StringVar(&c.TrustedSubnet, "t", c.TrustedSubnet, "Trusted CIDR subnet")
	flag.StringVar(&c.GRPCAddress, "g", ":3200", "Адрес запуска gRPC-сервера")
	flag.Parse()
	// Считаем переменные окружения более приоритетными перед флагами
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
	if _, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		c.EnableHTTPS = true
	}

	if trustedSubnet, ok := os.LookupEnv("TRUSTED_SUBNET"); ok {
		c.TrustedSubnet = trustedSubnet
	}
	if GRPCAddress, ok := os.LookupEnv("GRPC_ADDRESS"); ok {
		c.GRPCAddress = GRPCAddress
	}

	return c
}

// loadConfigFromFile чтение JSON файла конфигурации
func loadConfigFromFile() error {
	if ConfigPath == "" {
		log.Println("the path to the configuration file is empty")
		return nil
	}

	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err = json.Unmarshal(data, &Cfg); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}
	return nil
}
