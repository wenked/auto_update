package server

import (
	"auto-update/internal/database/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) CreateUser(c echo.Context) error {

	createUser := new(models.CreateUser)

	if err := c.Bind(createUser); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return nil
}
