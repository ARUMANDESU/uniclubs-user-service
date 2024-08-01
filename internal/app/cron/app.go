package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/ARUMANDESU/uniclubs-user-service/pkg/logger"
	"github.com/go-co-op/gocron/v2"
)

const (
	deleteNonActivatedUsersTime = time.Second * 10
	deleteNonActivatedUsersDays = 1
)

type App struct {
	log         *slog.Logger
	scheduler   gocron.Scheduler
	userStorage UserStorage
}

type UserStorage interface {
	DeleteNonActivatedUsers(ctx context.Context, days int) error
}

func New(log *slog.Logger, userStorage UserStorage) (*App, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &App{log: log, scheduler: s, userStorage: userStorage}, nil
}

func (a *App) Start() {
	a.log.Info("starting cron jobs")
	a.scheduler.NewJob(
		gocron.DurationJob(deleteNonActivatedUsersTime),
		gocron.NewTask(func() {
			ctx := context.Background()

			err := a.userStorage.DeleteNonActivatedUsers(ctx, deleteNonActivatedUsersDays)
			if err != nil {
				a.log.Error("failed to delete non activated users", logger.Err(err))
			}

			a.log.Debug("non activated users deleted")
		}),
	)
	a.scheduler.Start() // start is non-blocking so we don't need to start it in a goroutine
}
