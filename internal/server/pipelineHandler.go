package server

import (
	"auto-update/internal/database/models"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func (s *Server) CreatePipelineHandler(c echo.Context) error {
	name := c.FormValue("name")

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

	id, err := s.db.CreatePipeline(name, loggedUserId)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error creating pipeline",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message":     "ok",
		"pipeline_id": strconv.FormatInt(id, 10),
	})

}

func (s *Server) UpdatePipelineHandler(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

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

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid id",
		})
	}

	name := c.FormValue("name")

	updatePipeline := &models.UpdatePipeline{
		ID:   id,
		Name: name,
	}

	err = s.db.UpdatePipeline(updatePipeline, loggedUserId)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error updating pipeline",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "ok",
	})

}

func (s *Server) DeletePipelineHandler(c echo.Context) error {
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

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid id",
		})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid id",
		})
	}

	err = s.db.DeletePipeline(id, loggedUserId)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error deleting pipeline",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "ok",
	})

}

func (s *Server) ListPipelinesHandler(c echo.Context) error {
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

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid id",
		})
	}

	pipelines, err := s.db.ListPipelines(loggedUserId)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error getting pipelines",
		})
	}

	return c.JSON(http.StatusOK, pipelines)

}

func (s *Server) UpdateProdPipelineHandler(c echo.Context) error {
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

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid id",
		})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid id",
		})
	}

	userPipeline, err := s.db.GetUserPipelineById(id, loggedUserId)

	fmt.Println(userPipeline, "userPipeline")

	if err != nil {
		slog.Error("Pipeline not found", "error", err)

		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "user pipeline not found",
		})
	}

	go func() {
		s.sshclient.UpdateProductionNew(id, loggedUserId)
	}()

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Atualizaçãp de pipeline de produção iniciada com sucesso",
	})
}

func (s *Server) UpdateProductionById(c echo.Context) error {
	fmt.Println("Atualizando produção")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		fmt.Println("error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid id",
		})
	}

	go func() {
		s.sshclient.UpdateProductionById(id)
	}()

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Atualização de produção iniciada com sucesso",
	})

}
