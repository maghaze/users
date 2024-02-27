package grpc

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/maghaze/contracts/pbs/go/auth"
)

type AuthClient interface {
	GenerateToken(ctx context.Context, id uint64) (string, error)
}

type authClient struct {
	logger *zap.Logger
	api    pb.AuthClient
}

func NewAuthClient(cfg *Config, lg *zap.Logger) *authClient {
	client := &authClient{logger: lg}

	connection, err := grpc.Dial(cfg.Targets.Auth, grpc.WithInsecure())
	if err != nil {
		lg.Panic("error while instantiating auth grpc client", zap.Error(err))
	}
	client.api = pb.NewAuthClient(connection)

	return client
}

func (c *authClient) GenerateToken(ctx context.Context, id uint64) (string, error) {
	pbToken, err := c.api.CreateTokenFromId(ctx, &pb.Id{Value: id})
	if err != nil {
		errString := "Error generating token for given id"
		c.logger.Error(errString, zap.Uint64("id", id), zap.Error(err))
		return "", errors.New(errString)
	}
	return pbToken.Value, nil
}
