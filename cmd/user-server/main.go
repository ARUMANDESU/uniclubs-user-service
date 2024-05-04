package main

import (
	"context"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/app"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/config"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/logger"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		log.Error("error loading .env file")
	}

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Error("error loading aws config", logger.Err(err))
		panic(err)
	}

	log.Info("starting application",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.GRPC.Port),
	)

	application := app.New(log, cfg, awsCfg)
	go application.GRPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("stopping application", slog.String("signal", sign.String()))
	application.Close()
	log.Info("application stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
