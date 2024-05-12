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
	GRPCSrv  *grpcapp.App
	log      *slog.Logger
	postgres *postgresql.Storage
	redis    *redis.Storage
	rabbitMQ *rabbitmq.Rabbitmq
}

func New(log *slog.Logger, cfg *config.Config) *App {
	const op = "app.new"
	l := log.With(slog.String("op", op))

	postgres, err := postgresql.New(cfg.DatabaseDSN)
	if err != nil {
		l.Error("failed to connect to postgresql", logger.Err(err))
		panic(err)
	}
	redisStorage, err := redis.New(cfg.RedisURL)
	if err != nil {
		l.Error("failed to connect to redis", logger.Err(err))
		panic(err)
	}

	rabbitMQ, err := rabbitmq.New(cfg.Rabbitmq)
	if err != nil {
		l.Error("failed to connect to rabbitmq", logger.Err(err))
		panic(err)
	}

	authService := auth.New(log, cfg.Jwt, postgres, redisStorage, redisStorage, rabbitMQ)
	managementService := management.New(log, postgres, rabbitMQ)

	grpcApp := grpcapp.New(log, cfg.GRPC.Port, authService, managementService)

	return &App{GRPCSrv: grpcApp, log: log, postgres: postgres, redis: redisStorage, rabbitMQ: rabbitMQ}
}

func (a *App) Close() {
	const op = "app.close"
	log := a.log.With(slog.String("op", op))

	a.GRPCSrv.Stop()

	err := a.postgres.Close()
	if err != nil {
		log.Warn(op, "failed to close postgres connection", logger.Err(err))
	}

	err = a.redis.Close()
	if err != nil {
		log.Warn(op, "failed to close redis connection", logger.Err(err))
	}

	err = a.rabbitMQ.Close()
	if err != nil {
		log.Warn(op, "failed to close rabbitmq connection", logger.Err(err))
	}
}
