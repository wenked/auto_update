package server

import (
	"auto-update/internal/database/models"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func getLoggedUserId(loggedUser *jwt.Token) (int64, error) {
	fmt.Println(loggedUser, "loggedUser")

	claims, ok := loggedUser.Claims.(jwt.MapClaims)

	if !ok {
		return 0, errors.New("Error in get user claims")
	}

	idValue, ok := claims["id"]

	if !ok {
		return 0, errors.New("Error geting user id")
	}
	fmt.Println(idValue, "idValue")

	idInt, ok := idValue.(float64)

	if !ok {
		return 0, errors.New("Error converting user id to float64")
	}

	id := int64(idInt)
	return id, nil
}

func generateHashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (s *Server) CreateUserHandler(c echo.Context) error {

	createUser := new(models.CreateUser)

	if err := c.Bind(createUser); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	createUserSecret := os.Getenv("CREATE_USER_KEY")

	if createUserSecret != createUser.CreateUserKey {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid create user key"})
	}

	userExists, _ := s.db.GetUserByEmail(createUser.Email)

	if userExists.ID != 0 {
		slog.Error("User with this email already exists")

		return c.JSON(http.StatusBadRequest, map[string]string{"message": "user with this email already exists"})
	}

	password, err := generateHashPassword(createUser.Password)

	if err != nil {
		slog.Error("error generating password hash", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	id, err := s.db.CreateUser(createUser.Name, createUser.Email, password)

	if err != nil {
		slog.Error("error creating user", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]int64{"id": id})

}

func (s *Server) UpdateUserHandler(c echo.Context) error {
	stringID := c.Param("id")
	loggedUser, ok := c.Get("user").(*jwt.Token)

	if !ok {
		slog.Error("Error getting logged user context")
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}

	loggedUserId, err := getLoggedUserId(loggedUser)

	if err != nil {
		slog.Error("Error getting logged user id ")
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}

	id, err := strconv.ParseInt(stringID, 10, 64)

	if err != nil {
		slog.Error("error parsing id", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})

	}

	if loggedUserId != id {
		slog.Error("invalid user id")
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid user id"})
	}

	updateUserRequest := new(models.UpdateUserRequest)

	if err := c.Bind(updateUserRequest); err != nil {
		slog.Error("error updating user", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request fields"})
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

func (s *Server) DeleteUserHandler(c echo.Context) error {
	userID := c.Param("id")

	id, err := strconv.ParseInt(userID, 10, 64)

	if err != nil {
		slog.Error("error parsing id", err)

		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	err = s.db.DeleteUser(id)

	if err != nil {
		slog.Error("error deleting user", err)

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "user deleted"})

}

func (s *Server) ListUsersHandler(c echo.Context) error {
	page := c.Param("page")

	pageInt, err := strconv.ParseInt(page, 10, 64)

	if err != nil {
		pageInt = 1
	}

	limit := c.Param("limit")

	limitInt, err := strconv.ParseInt(limit, 10, 64)

	if err != nil {
		limitInt = 50
	}

	users, err := s.db.ListUsers(pageInt, limitInt)

	if err != nil {
		slog.Error("error listing users", err)

		return c.JSON(http.StatusBadRequest, map[string]string{"message": "error listing users"})
	}

	return c.JSON(http.StatusOK, users)

}

func (s *Server) CreateNotificationConfigHandler(c echo.Context) error {
	loggedUser, ok := c.Get("user").(*jwt.Token)

	if !ok {
		slog.Error("Error getting logged user context")
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}

	loggedUserId, err := getLoggedUserId(loggedUser)

	if err != nil {
		slog.Error("Error getting logged user id ")
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}

	fmt.Println(loggedUserId)

	newNotificationConfig := new(models.NotificationConfig)

	if err := c.Bind(newNotificationConfig); err != nil {
		slog.Error("error binding notificationConfig user", "err", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request fields"})
	}

	newNotificationConfig.UserID = loggedUserId

	id, err := s.db.CreateNotificationConfig(newNotificationConfig)

	if err != nil {
		slog.Error("Error creating notification conifg", "error", err)

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error creating notificationConfig",
		})
	}

	strId := strconv.FormatInt(id, 10)
	return c.JSON(http.StatusOK, map[string]string{
		"id": strId,
	})
}
