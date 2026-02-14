package config

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPConfig struct {
	Address string        `yaml:"address" env:"API_ADDRESS" env-default:":8080"`
	Timeout time.Duration `yaml:"timeout" env:"API_TIMEOUT" env-default:"5s"`
}

type Config struct {
	LogLevel     string     `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	HTTP         HTTPConfig `yaml:"api_server"`
	TasksAddress string     `yaml:"tasks_address" env:"TASKS_ADDRESS" env-default:"tasks:8080"`
}

func MustLoad(configPath string) Config {
	var cfg Config

	// если путь пустой - просто env
	if configPath == "" {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("cannot read env: %s", err)
		}
		return cfg
	}

	// пробуем файл, если его нет - env
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		var pe *os.PathError
		if errors.As(err, &pe) {
			if err := cleanenv.ReadEnv(&cfg); err != nil {
				log.Fatalf("cannot read env: %s", err)
			}
			return cfg
		}
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}

	return cfg
}
