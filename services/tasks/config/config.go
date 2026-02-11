package config

import (
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type Config struct {
	LogLevel  string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	Address   string `yaml:"tasks_address" env:"TASKS_ADDRESS" env-default:":8080"`
	DBAddress string `yaml:"db_address" env:"DB_ADDRESS" env-required:"true"`
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
