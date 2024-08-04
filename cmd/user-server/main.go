package main

import (
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ARUMANDESU/uniclubs-user-service/internal/app"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		log.Error("error loading .env file")
	}

	cfg := config.MustLoad()
	log := logger.SetupLogger(cfg.Env)

	log.Info("starting application",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.GRPC.Port),
	)

	application := app.New(log, cfg)

	go application.GRPCSrv.MustRun()
	application.CronJobs.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("stopping application", slog.String("signal", sign.String()))
	application.Close()
	log.Info("application stopped")
}
