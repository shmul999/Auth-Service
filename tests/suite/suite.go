package suite

import (
	"context"
	"net"
	"sso/internal/config"
	"strconv"
	"testing"

	ssov1 "github.com/shmul999/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct{
	*testing.T
	Cfg *config.Config
	AuthClient ssov1.AuthClient //Клиент для взаимодействия с gRPC сервером
}

const(
	grpcHost = "localhost"
)

func New(t *testing.T) (context.Context, *Suite) {
    t.Helper()
    t.Parallel()

    cfg := config.MustLoadByPath("../config/local.yaml")

    ctx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
    
    t.Cleanup(func() {
        t.Helper()
        cancel()
    })

    cc, err := grpc.NewClient(
        grpcAddress(cfg),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        t.Fatalf("grpc client creation failed: %v", err)
    }

    authClient := ssov1.NewAuthClient(cc)

    return ctx, &Suite{
        T:          t,
        Cfg:        cfg,
        AuthClient: authClient,
    }
}

func grpcAddress(cfg *config.Config) string{
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}