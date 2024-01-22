package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env         string `yaml:"env" env:"ENV" env-default:"local"`
	GRPC        GRPC   `yaml:"grpc"`
	DatabaseDSN string `yaml:"database_dsn" env:"DATABASE_DSN" env-required:"true"`
	RedisURL    string `yaml:"redis_url" env:"REDIS_URL" env-required:"true"`
}

type GRPC struct {
	Port    int           `yaml:"port" env:"GRPC_PORT"`
	Timeout time.Duration `yaml:"timeout" env:"GRPC_TIMEOUT"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(configPath string) *Config {
	cfg, err := LoadByPath(configPath)
	if err != nil {
		panic(err)
	}

	return cfg
}

func LoadByPath(configPath string) (*Config, error) {
	var cfg Config

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err = cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("failed to read env: %w", err)
		}
		return &cfg, nil
	}

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "./config/local.yaml", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
