package http

import (
	"encoding/json"
	"fmt"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/maghaze/users/internal/ports/grpc"
	"github.com/maghaze/users/internal/repository"
	"go.uber.org/zap"
)

type Server struct {
	logger     *zap.Logger
	auth       grpc.AuthClient
	repository repository.Repository

	managementApp *fiber.App
	clientApp     *fiber.App
}

func New(log *zap.Logger, repo repository.Repository, auth grpc.AuthClient) *Server {
	server := &Server{logger: log, repository: repo, auth: auth}

	fiberConfig := fiber.Config{JSONEncoder: json.Marshal, JSONDecoder: json.Unmarshal, DisableStartupMessage: true}
	server.managementApp, server.clientApp = fiber.New(fiberConfig), fiber.New(fiberConfig)

	prometheus := fiberprometheus.New("users")
	prometheus.RegisterAt(server.managementApp, "/metrics")
	server.managementApp.Use(prometheus.Middleware)

	server.managementApp.Get("/healthz/liveness", server.liveness)
	server.managementApp.Get("/healthz/readiness", server.readiness)

	v1 := server.clientApp.Group("/v1")
	v1.Post("/register", server.register)
	v1.Post("/login", server.login)
	// v1.Get("/:id<uint64>", server.fetchUserId, server.user)
	// v1.Get("/me", server.fetchUserId, server.me)
	// v1.Post("/update-information", server.fetchUserId, server.updateInformation)
	// v1.Post("/update-password", server.fetchUserId, server.updatePassword)

	return server
}

func (server *Server) Serve(managementPort, clientPort int) {
	go func() {
		server.logger.Info("HTTP management server starts listening on", zap.Int("port", managementPort))
		if err := server.managementApp.Listen(fmt.Sprintf(":%d", managementPort)); err != nil {
			server.logger.Fatal("error resolving HTTP server", zap.Error(err))
		}
	}()

	go func() {
		server.logger.Info("HTTP client server starts listening on", zap.Int("port", clientPort))
		if err := server.clientApp.Listen(fmt.Sprintf(":%d", clientPort)); err != nil {
			server.logger.Fatal("error resolving HTTP server", zap.Error(err))
		}
	}()
}
