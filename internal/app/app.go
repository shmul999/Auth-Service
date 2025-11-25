package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/services/auth"
	"sso/internal/storage/sqlite"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, TokenTTL time.Duration) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authSercice := auth.New(log, storage, storage, TokenTTL, storage)
	grpcApp := grpcapp.New(log, grpcPort, authSercice)

	return &App{
		GRPCSrv: grpcApp,
	}
}
