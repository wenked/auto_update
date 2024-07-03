package server

import (
	"auto-update/internal/database/models"
	"auto-update/utils"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (s *Server) CreateServerHandler(c echo.Context) error {
	fmt.Println("Criando servidor")

	serverinfo := new(ServerInfo)

	if err := c.Bind(serverinfo); err != nil {
		slog.Error("Error binding body", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid request",
		})
	}

	fmt.Println("serverinfo", serverinfo.Host)
	fmt.Println("serverinfo", serverinfo.Password)
	fmt.Println("serverinfo", serverinfo.Script)
	fmt.Println("serverinfo", serverinfo.PipelineID)
	fmt.Println("serverinfo", serverinfo)

	hashedPassword, err := utils.Encrypt(serverinfo.Password)

	if err != nil {
		slog.Error("Error hashing password", "error", err)

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error creating server",
		})
	}

	newId, err := s.db.CreateServer(serverinfo.Host, hashedPassword, serverinfo.Script, serverinfo.PipelineID, serverinfo.Label)

	if err != nil {
		slog.Error("Error creating server", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error creating server",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message":   "ok",
		"server_id": strconv.FormatInt(newId, 10),
	})
}

func (s *Server) UpdateServerHandler(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		slog.Error("Invalid id", "error", err)

		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid id",
		})
	}

	serverinfo := new(ServerInfo)

	if err := c.Bind(serverinfo); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid request",
		})
	}

	if serverinfo.Password != "" {
		serverinfo.Password, err = generateHashPassword(serverinfo.Password)

		if err != nil {
			slog.Error("Error hashing password", "error ", err)
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Internal server error",
			})
		}
	}

	updateServer := &models.UpdateServer{
		ID:         id,
		Host:       serverinfo.Host,
		Password:   serverinfo.Password,
		Script:     serverinfo.Script,
		Label:      serverinfo.Label,
		PipelineID: serverinfo.PipelineID,
		Active:     serverinfo.Active,
	}

	err = s.db.UpdateServer(updateServer)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error updating server",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "ok",
	})

}

func (s *Server) DeleteServerHandler(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid id",
		})
	}

	err = s.db.DeleteServer(id)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error deleting server",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "ok",
	})

}

func (s *Server) ListServersHandler(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	servers, err := s.db.ListServers(id)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error getting servers",
		})
	}

	return c.JSON(http.StatusOK, servers)
}

func (s *Server) UpdatePasswords(c echo.Context) error {

	err := s.db.UpdateServersPasswords()

	if err != nil {
		slog.Error("error updating passowrds", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "internal server error",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "passwords updated",
	})
}
