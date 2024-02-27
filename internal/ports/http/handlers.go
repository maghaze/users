package http

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/maghaze/users/internal/models"
	"github.com/maghaze/users/pkg/rdbms"
	"go.uber.org/zap"
)

func (handler *Server) liveness(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusOK)
}

func (handler *Server) readiness(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusOK)
}

func (handler *Server) register(c *fiber.Ctx) error {
	ctx := c.Context()

	request := struct{ Email, Password string }{}
	if err := c.BodyParser(&request); err != nil {
		errString := "Error parsing request body"
		handler.logger.Error(errString, zap.Any("request", request), zap.Error(err))
		return c.Status(http.StatusBadRequest).SendString(errString)
	}

	user, err := handler.repository.FindUserByEmail(ctx, request.Email)
	if err != nil && err.Error() != rdbms.ErrReadNotFound {
		errString := "Error while retrieving data from database"
		handler.logger.Error(errString, zap.Error(err))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	} else if err == nil || (user != nil && user.Id != 0) {
		errString := "User with given email already exists"
		handler.logger.Error(errString, zap.String("email", request.Email))
		return c.Status(http.StatusBadRequest).SendString(errString)
	}

	user = &models.User{Email: request.Email, Password: request.Password}
	if err := handler.repository.CreateUser(ctx, user); err != nil {
		errString := "Error happened while creating the user"
		handler.logger.Error(errString, zap.Error(err))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	if user.Id == 0 {
		errString := "Error invalid user id created"
		handler.logger.Error(errString, zap.Any("user", user))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	token, err := handler.auth.GenerateToken(ctx, user.Id)
	if err != nil {
		errString := "Error creating JWT token for user"
		handler.logger.Error(errString, zap.Any("user", user), zap.Error(err))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	response := map[string]string{"Token": token}
	return c.Status(http.StatusCreated).JSON(&response)
}

func (handler *Server) login(c *fiber.Ctx) error {
	ctx := c.Context()

	request := struct{ Email, Password string }{}
	if err := c.BodyParser(&request); err != nil {
		errString := "Error parsing request body"
		handler.logger.Error(errString, zap.Error(err))
		return c.Status(http.StatusBadRequest).SendString(errString)
	}

	user, err := handler.repository.FindUserByEmailAndPassword(ctx, request.Email, request.Password)
	if err != nil {
		errString := "Wrong email or password has been given"
		handler.logger.Error(errString, zap.Error(err))
		return c.Status(http.StatusBadRequest).SendString(errString)
	}

	if user == nil {
		errString := "Error invalid user returned"
		handler.logger.Error(errString, zap.Any("request", request))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	// request token
	token, err := handler.auth.GenerateToken(ctx, user.Id)
	if err != nil {
		errString := "Error creating JWT token for user"
		handler.logger.Error(errString, zap.Any("user", user), zap.Error(err))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	response := map[string]string{"Token": token}
	return c.Status(http.StatusOK).JSON(&response)
}

// get user by id
func (handler *Server) user(c *fiber.Ctx) error {
	ctx := c.Context()
	idString := c.Params("id")

	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		errString := "Error invalid id for the user"
		handler.logger.Error(errString, zap.String("id", idString))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	} else if id <= 0 {
		errString := "Error invalid id has been given"
		handler.logger.Error(errString)
		return c.Status(http.StatusBadRequest).SendString(errString)
	}

	user, err := handler.repository.FindUserById(ctx, id)
	if err != nil {
		if err.Error() == rdbms.ErrReadNotFound {
			errString := "User with given id doesn't exists"
			return c.Status(http.StatusBadRequest).SendString(errString)
		}

		errString := "Error while retrieving the user"
		handler.logger.Error(errString, zap.Error(err))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	if user == nil {
		errString := "Error invalid user returned"
		handler.logger.Error(errString, zap.Uint64("id", id))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	return c.Status(http.StatusOK).JSON(&user)
}

// get user of the header
func (handler *Server) me(c *fiber.Ctx) error {
	ctx := c.Context()

	id, ok := c.Locals("id").(uint64)
	if !ok {
		errString := "Error invalid id for the user"
		handler.logger.Error(errString, zap.Any("id", c.Locals("id")))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	} else if id <= 0 {
		errString := "Error invalid id has been given"
		handler.logger.Error(errString)
		return c.Status(http.StatusBadRequest).SendString(errString)
	}

	user, err := handler.repository.FindUserById(ctx, id)
	if err != nil {
		if err.Error() == rdbms.ErrReadNotFound {
			errString := "Error finding user of the request"
			return c.Status(http.StatusInternalServerError).SendString(errString)
		}

		errString := "Error while retrieving the user"
		handler.logger.Error(errString, zap.Error(err))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	if user == nil {
		errString := "Error invalid user returned"
		handler.logger.Error(errString, zap.Any("id", id))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	return c.Status(http.StatusOK).JSON(&user)
}

func (handler *Server) updateInformation(c *fiber.Ctx) error {
	ctx := c.Context()

	id, ok := c.Locals("id").(uint64)
	if !ok {
		errString := "Error invalid id for the user"
		handler.logger.Error(errString, zap.Any("id", c.Locals("id")))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	} else if id <= 0 {
		errString := "Error invalid id has been given"
		handler.logger.Error(errString)
		return c.Status(http.StatusBadRequest).SendString(errString)
	}

	request := struct{ FirstName, LastName string }{}
	if err := c.BodyParser(&request); err != nil {
		errString := "Error parsing request body"
		handler.logger.Error(errString, zap.Error(err))
		return c.Status(http.StatusBadRequest).SendString(errString)
	}

	if len(request.FirstName) == 0 && len(request.LastName) == 0 {
		errString := "An empty request body has been given"
		handler.logger.Error(errString)
		return c.Status(http.StatusBadRequest).SendString(errString)
	}

	user, err := handler.repository.FindUserById(ctx, id)
	if err != nil {
		if err.Error() == rdbms.ErrReadNotFound {
			errString := "Error finding user of the request"
			return c.Status(http.StatusInternalServerError).SendString(errString)
		}

		errString := "Error while retrieving the user"
		handler.logger.Error(errString, zap.Error(err))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	if user == nil {
		errString := "Error invalid user returned"
		handler.logger.Error(errString, zap.Any("id", id))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	if len(request.FirstName) != 0 {
		user.FirstName = request.FirstName
	}

	if len(request.LastName) != 0 {
		user.LastName = request.LastName
	}

	if err := handler.repository.UpdateUser(ctx, user); err != nil {
		errString := "Error while updating the user"
		handler.logger.Error(errString, zap.Error(err))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	return c.SendStatus(http.StatusOK)
}

func (handler *Server) updatePassword(c *fiber.Ctx) error {
	ctx := c.Context()

	id, ok := c.Locals("id").(uint64)
	if !ok {
		errString := "Error invalid id for the user"
		handler.logger.Error(errString, zap.Any("id", c.Locals("id")))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	} else if id <= 0 {
		errString := "Error invalid id has been given"
		handler.logger.Error(errString)
		return c.Status(http.StatusBadRequest).SendString(errString)
	}

	request := struct{ OldPassword, NewPassword string }{}
	if err := c.BodyParser(&request); err != nil {
		errString := "Error parsing request body"
		handler.logger.Error(errString, zap.Error(err))
		return c.Status(http.StatusBadRequest).SendString(errString)
	}

	if len(request.OldPassword) == 0 {
		errString := "Invalid old password has been given"
		handler.logger.Error(errString)
		return c.Status(http.StatusBadRequest).SendString(errString)
	} else if len(request.NewPassword) == 0 {
		errString := "Invalid password has been given"
		handler.logger.Error(errString)
		return c.Status(http.StatusBadRequest).SendString(errString)
	}

	user, err := handler.repository.FindUserById(ctx, id)
	if err != nil {
		if err.Error() == rdbms.ErrReadNotFound {
			errString := "Error finding user of the request"
			return c.Status(http.StatusInternalServerError).SendString(errString)
		}

		errString := "Error while retrieving the user"
		handler.logger.Error(errString, zap.Error(err))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	if user == nil {
		errString := "Error invalid user returned"
		handler.logger.Error(errString, zap.Any("id", id))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	if request.OldPassword != user.Password {
		errString := "Error wrong old password"
		handler.logger.Error(errString, zap.Uint64("id", id), zap.Any("request", request))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	} else {
		user.Password = request.NewPassword
	}

	if err := handler.repository.UpdateUser(ctx, user); err != nil {
		errString := "Error while updating the user"
		handler.logger.Error(errString, zap.Error(err))
		return c.Status(http.StatusInternalServerError).SendString(errString)
	}

	return c.SendStatus(http.StatusOK)
}
