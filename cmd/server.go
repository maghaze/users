package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/maghaze/users/internal/config"
	"github.com/maghaze/users/internal/ports/grpc"
	"github.com/maghaze/users/internal/ports/http"
	"github.com/maghaze/users/internal/repository"
	"github.com/maghaze/users/pkg/logger"
	"github.com/maghaze/users/pkg/rdbms"
)

type Server struct {
	managementPort int
	clientPort     int
}

func (server Server) Command(trap chan os.Signal) *cobra.Command {
	run := func(_ *cobra.Command, _ []string) {
		server.main(config.Load(true), trap)
	}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "serve api-gateway server",
		Run:   run,
	}

	cmd.Flags().IntVar(&server.managementPort, "management-port", 8020, "The port the metrics and probe endpoints binds to")
	cmd.Flags().IntVar(&server.clientPort, "client-port", 8021, "The port the api-gateway http server endpoints binds to")

	return cmd
}

func (server *Server) main(cfg *config.Config, trap chan os.Signal) {
	logger := logger.NewZap(cfg.Logger)

	rdbms, err := rdbms.NewPostgres(cfg.RDBMS)
	if err != nil {
		logger.Panic("Error creating rdbms database", zap.Error(err))
	}

	repo := repository.New(logger, rdbms)
	authGrpcClient := grpc.NewAuthClient(cfg.GRPC, logger)

	go http.New(logger, repo, authGrpcClient).Serve(server.managementPort, server.clientPort)

	// Keep this at the bottom of the main function
	field := zap.String("signal trap", (<-trap).String())
	logger.Info("exiting by receiving a unix signal", field)
}
