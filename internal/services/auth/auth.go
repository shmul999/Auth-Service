package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct{
	log *slog.Logger
	usrSaver UserSaver
	usrProvider UserProvider
	AppProvider AppProvider
	tokenTTl time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface{
	User (ctx context.Context, email string) (models.User, error)
	IsAdmin (ctx context.Context, UserID int64) (bool, error)
}

type AppProvider interface{
	App(ctx context.Context, appID int) (models.App, error)
}

var(
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists = errors.New("user exists")
)
func New(log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	tokenTTL time.Duration,
	appProvider AppProvider,
)*Auth{
	return &Auth{
			log: log,
			usrSaver: userSaver,
			usrProvider: userProvider,
			tokenTTl: tokenTTL,
			AppProvider: appProvider,
	}
}

func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error){
	const op = "auth.Login"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("logging user ...")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil{
		if errors.Is(err, storage.ErrUserNotFound){
			log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		log.Error("failed to get user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil{
		log.Info("invalid credentials")

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}
	app, err := a.AppProvider.App(ctx, appID)
	if err != nil{
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.NewToken(user, app, a.tokenTTl)
	if err != nil{
		log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user logged in successfully")
	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error){
	const op = "auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("registering user ...")

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil{
		log.Error("failed to generate password hash", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, hashedPass)
	if err != nil{
		if errors.Is(err, storage.ErrUserExists){
			log.Warn("user already exists", sl.Err(err))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")
	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error){
	const op = "auth.IsAdmin"
	log := a.log.With(
		slog.String("op", op),
		slog.Int64("userID", userID),
	)

	log.Info("checking if user is admin")

	isAdm, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil{
		if errors.Is(err, storage.ErrAppnotFound){
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdm))

	return isAdm, nil
}