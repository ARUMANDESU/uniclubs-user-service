package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestLoadByPath_HappyPath(t *testing.T) {
	fileName, err := createTempConfigFile()
	if err != nil {
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		os.Remove(fileName)
	})

	cfg, err := LoadByPath(fileName)

	require.NoError(t, err)
	assertConfig(t, cfg)
}

func TestLoadFromEnv_HappyPath(t *testing.T) {
	setEnvVariables()
	t.Cleanup(func() {
		unsetEnvVariables()
	})

	cfg := MustLoadFromEnv()

	assertConfig(t, cfg)
}

func TestLoadFromEnv_FailPath(t *testing.T) {
	//setEnvVariables()

	require.Panics(t, func() {
		MustLoadFromEnv()
	})

}

func TestMustLoad_FromYaml(t *testing.T) {
	fileName, err := createTempConfigFile()
	if err != nil {
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		os.Remove(fileName)
	})

	os.Args = []string{"cmd", "-config", fileName}
	cfg := MustLoad()

	assertConfig(t, cfg)
}

func TestMustLoad_FromEnv(t *testing.T) {
	setEnvVariables()
	t.Cleanup(func() {
		unsetEnvVariables()
	})

	cfg := MustLoad()

	assertConfig(t, cfg)
}

func TestMustLoad_FromYaml_FailPath(t *testing.T) {
	os.Args = []string{"cmd", "-config", "./some/wrong/path"}

	require.Panics(t, func() {
		MustLoad()
	})

}

func TestMustLoad_FromEnv_FailPath(t *testing.T) {
	os.Args = []string{"cmd", "-config", ""}

	require.Panics(t, func() {
		MustLoad()
	})

}

func TestLoadByPath_FailPath_WrongPath(t *testing.T) {
	_, err := LoadByPath("./some/wrong/path")

	require.Error(t, err, "have to return error")
}

func createTempConfigFile() (string, error) {
	f, err := os.Create("./temp_config.yaml")
	if err != nil {
		return "", err
	}
	defer f.Close()

	configData := []byte(`env: "test"
database_dsn: "test_database_dsn"
redis_url: "test_redis_url"
grpc:
  port: 44044
  timeout: "1h"
rabbitmq:
  user: "admin"
  password: "admin"
  host: "localhost"
  port: "5672"
  exchange_name: "test_exchange"
  queue_name: "test_queue"
clients:
  image:
    address: "image_service_address"
    timeout: "3s"
    retries_count: 3`)

	_, err = f.Write(configData)
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}

func setEnvVariables() {
	os.Setenv("ENV", "test")
	os.Setenv("DATABASE_DSN", "test_database_dsn")
	os.Setenv("REDIS_URL", "test_redis_url")
	os.Setenv("GRPC_PORT", "44044")
	os.Setenv("GRPC_TIMEOUT", "1h")
	os.Setenv("RABBITMQ_USER", "admin")
	os.Setenv("RABBITMQ_PASSWORD", "admin")
	os.Setenv("RABBITMQ_HOST", "localhost")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_EXCHANGE_NAME", "test_exchange")
	os.Setenv("RABBITMQ_QUEUE_NAME", "test_queue")
	os.Setenv("IMAGE_SERVICE_ADDRESS", "image_service_address")
	os.Setenv("IMAGE_SERVICE_TIMEOUT", "3s")
	os.Setenv("IMAGE_SERVICE_RETRIES_COUNT", "3")
}

func assertConfig(t *testing.T, cfg *Config) {
	t.Helper()

	assert.Equal(t, "test", cfg.Env)
	assert.Equal(t, "test_database_dsn", cfg.DatabaseDSN)
	assert.Equal(t, "test_redis_url", cfg.RedisURL)
	assert.Equal(t, 44044, cfg.GRPC.Port)
	assert.Equal(t, time.Hour, cfg.GRPC.Timeout)
	assert.Equal(t, "admin", cfg.Rabbitmq.User)
	assert.Equal(t, "admin", cfg.Rabbitmq.Password)
	assert.Equal(t, "localhost", cfg.Rabbitmq.Host)
	assert.Equal(t, "5672", cfg.Rabbitmq.Port)
	assert.Equal(t, "test_exchange", cfg.Rabbitmq.ExchangeName)
	assert.Equal(t, "test_queue", cfg.Rabbitmq.QueueName)
	assert.Equal(t, "image_service_address", cfg.Clients.Image.Address)
	assert.Equal(t, 3*time.Second, cfg.Clients.Image.Timeout)
	assert.Equal(t, 3, cfg.Clients.Image.RetriesCount)
}

func unsetEnvVariables() {
	os.Unsetenv("ENV")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("GRPC_TIMEOUT")
	os.Unsetenv("RABBITMQ_USER")
	os.Unsetenv("RABBITMQ_PASSWORD")
	os.Unsetenv("RABBITMQ_HOST")
	os.Unsetenv("RABBITMQ_PORT")
	os.Unsetenv("RABBITMQ_EXCHANGE_NAME")
	os.Unsetenv("RABBITMQ_QUEUE_NAME")
	os.Unsetenv("IMAGE_SERVICE_ADDRESS")
	os.Unsetenv("IMAGE_SERVICE_TIMEOUT")
	os.Unsetenv("IMAGE_SERVICE_RETRIES_COUNT")
}
