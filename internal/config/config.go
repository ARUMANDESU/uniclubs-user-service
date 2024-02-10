package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env         string        `yaml:"env" env:"ENV" env-default:"local"`
	GRPC        GRPC          `yaml:"grpc"`
	Rabbitmq    Rabbitmq      `yaml:"rabbitmq"`
	DatabaseDSN string        `yaml:"database_dsn" env:"DATABASE_DSN" env-required:"true"`
	RedisURL    string        `yaml:"redis_url" env:"REDIS_URL" env-required:"true"`
	Clients     ClientsConfig `yaml:"clients"`
}

type GRPC struct {
	Port    int           `yaml:"port" env:"GRPC_PORT"`
	Timeout time.Duration `yaml:"timeout" env:"GRPC_TIMEOUT"`
}

type Rabbitmq struct {
	User         string `yaml:"user" env:"RABBITMQ_USER"`
	Password     string `yaml:"password" env:"RABBITMQ_PASSWORD"`
	Host         string `yaml:"host" env:"RABBITMQ_HOST"`
	Port         string `yaml:"port" env:"RABBITMQ_PORT"`
	ExchangeName string `yaml:"exchange_name" env:"RABBITMQ_EXCHANGE_NAME"`
	QueueName    string `yaml:"queue_name" env:"RABBITMQ_QUEUE_NAME"`
}

type ClientsConfig struct {
	Image struct {
		Address      string        `yaml:"address" env:"IMAGE_SERVICE_ADDRESS"`
		Timeout      time.Duration `yaml:"timeout" env:"IMAGE_SERVICE_TIMEOUT"`
		RetriesCount int           `yaml:"retries_count" env:"IMAGE_SERVICE_RETRIES_COUNT"`
	} `yaml:"image"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		return MustLoadFromEnv()
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
		return nil, fmt.Errorf("there is no config file: %w", err)
	}

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}

func MustLoadFromEnv() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic("Env empty")
	}
	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
