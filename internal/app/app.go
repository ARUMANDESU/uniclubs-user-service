package app

import (
	grpcapp "github.com/ARUMANDESU/uniclubs-user-service/internal/app/grpc"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/config"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/services/auth"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage/postgresql"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {

	storage, err := postgresql.New(cfg.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	authService := auth.New(log, storage)

	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port)

	return &App{GRPCSrv: grpcApp}
}
