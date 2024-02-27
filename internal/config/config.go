package config

import (
	"github.com/maghaze/users/internal/ports/grpc"
	"github.com/maghaze/users/pkg/logger"
	"github.com/maghaze/users/pkg/rdbms"
)

type Config struct {
	Logger *logger.Config `koanf:"logger"`
	RDBMS  *rdbms.Config  `koanf:"rdbms"`
	GRPC   *grpc.Config   `koanf:"grpc"`
}
