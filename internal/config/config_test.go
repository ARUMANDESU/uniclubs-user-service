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

	cfg, err := loadByPath(fileName)

	require.NoError(t, err)
	assertConfig(t, cfg)
}

func TestLoadFromEnv_HappyPath(t *testing.T) {
	setEnvVariables()
	t.Cleanup(func() {
		unsetEnvVariables()
	})

	cfg := mustLoadFromEnv()

	assertConfig(t, cfg)
}

func TestLoadFromEnv_FailPath(t *testing.T) {
	//setEnvVariables()

	require.Panics(t, func() {
		mustLoadFromEnv()
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
	_, err := loadByPath("./some/wrong/path")

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
jwt:
  access_token_duration: "15m"
  access_token_secret: "secret"
  refresh_token_secret: "secret"
aws:
  region: "us-east-1"
  bucket: "bucket"`)

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
	os.Setenv("ACCESS_TOKEN_DURATION", "15m")
	os.Setenv("ACCESS_TOKEN_SECRET", "secret")
	os.Setenv("REFRESH_TOKEN_SECRET", "secret")
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_ACCESS_KEY_ID", "key_id")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "access_key")
	os.Setenv("AWS_S3_BUCKET", "bucket")
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
	assert.Equal(t, 15*time.Minute, cfg.Jwt.AccessTokenDuration)
	assert.Equal(t, "secret", cfg.Jwt.AccessTokenSecret)
	assert.Equal(t, "secret", cfg.Jwt.RefreshTokenSecret)
	assert.Equal(t, "us-east-1", cfg.AWS.Region)
	assert.Equal(t, "bucket", cfg.AWS.Bucket)
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
	os.Unsetenv("ACCESS_TOKEN_DURATION")
	os.Unsetenv("ACCESS_TOKEN_SECRET")
	os.Unsetenv("REFRESH_TOKEN_SECRET")
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_S3_BUCKET")
}
