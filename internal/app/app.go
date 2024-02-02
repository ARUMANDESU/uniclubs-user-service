package app

import (
	grpcapp "github.com/ARUMANDESU/uniclubs-user-service/internal/app/grpc"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/config"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/rabbitmq"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/services/auth"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/services/management"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage/postgresql"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage/redis"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/logger"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {
	const op = "App.New"
	l := log.With(slog.String("op", op))

	postgres, err := postgresql.New(cfg.DatabaseDSN)
	if err != nil {
		l.Error("failed to connect to postgresql", logger.Err(err))
		panic(err)
	}
	redisStrg, err := redis.New(cfg.RedisURL)
	if err != nil {
		l.Error("failed to connect to redis", logger.Err(err))
		panic(err)
	}

	rmq, err := rabbitmq.New(cfg.Rabbitmq)
	if err != nil {
		l.Error("failed to connect to rabbitmq", logger.Err(err))
		panic(err)
	}

	authService := auth.New(log, postgres, redisStrg, redisStrg, rmq)
	managementService := management.New(log, postgres)

	grpcApp := grpcapp.New(log, cfg.GRPC.Port, authService, managementService)

	return &App{GRPCSrv: grpcApp}
}
