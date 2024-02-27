package http

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (middleware *Server) fetchUserId(c *fiber.Ctx) error {
	header := c.Request().Header.Peek("X-User-Id")

	id, err := strconv.ParseUint(string(header), 10, 64)
	if err != nil {
		middleware.logger.Error("invalid id header", zap.ByteString("header", header), zap.Error(err))
		return err
	}

	if id == 0 {
		errString := "Invalid id value"
		middleware.logger.Error(errString, zap.ByteString("header", header))
		return errors.New(errString)
	}

	c.Locals("id", id)

	return c.Next()
}
