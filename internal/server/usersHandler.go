package server

import (
	"auto-update/internal/database/models"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func generateHashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (s *Server) CreateUser(c echo.Context) error {

	createUser := new(models.CreateUser)

	if err := c.Bind(createUser); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	createUserSecret := os.Getenv("CREATE_USER_KEY")

	if createUserSecret != createUser.CreateUserKey {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid create user key"})
	}

	password, err := generateHashPassword(createUser.Password)

	if err != nil {
		slog.Error("error generating password hash", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	id, err := s.db.CreateUser(createUser.Name, password, createUser.Email)

	if err != nil {
		slog.Error("error creating user", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]int64{"id": id})

}

func (s *Server) UpdateUser(c echo.Context) error {
	stringID := c.Param("id")

	id, err := strconv.ParseInt(stringID, 10, 64)

	if err != nil {
		slog.Error("error parsing id", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})

	}

	updateUserRequest := new(models.UpdateUserRequest)

	if err := c.Bind(updateUserRequest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	updateUser := &models.User{
		ID: id,
	}

	if updateUserRequest.Name != "" {
		updateUser.Name = updateUserRequest.Name
	}

	if updateUserRequest.Email != "" {
		updateUser.Email = updateUserRequest.Email
	}

	if updateUserRequest.Password != "" {
		password, err := generateHashPassword(updateUserRequest.Password)

		if err != nil {
			slog.Error("error generating password hash", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		updateUser.Password = password

	}

	if err != nil {
		slog.Error("error generating password hash", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	err = s.db.UpdateUser(updateUser)

	if err != nil {
		slog.Error("error updating user", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "user updated"})
}
