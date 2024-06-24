package server

import (
	"auto-update/internal/database/models"
	"auto-update/internal/sse"
	"auto-update/internal/sshclient"
	"auto-update/utils"
	"auto-update/views"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ServerInfo struct {
	Host       string `json:"host"`
	Password   string `json:"password"`
	Script     string `json:"script"`
	PipelineID int64  `json:"pipeline_id"`
	Label      string `json:"label"`
	Active     bool   `json:"active"`
}

type GithubWebhook struct {
	Ref         string      `json:"ref"`
	Pusher      Pusher      `json:"pusher"`
	HeadCommit  HeadCommit  `json:"head_commit"`
	Action      string      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
}

type HeadCommit struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}
type Pusher struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
type MergedBy struct {
	Login string `json:"login"`
}

type Repo struct {
	FullName string `json:"full_name"`
}

type Head struct {
	Ref  string `json:"ref"`
	Repo Repo   `json:"repo"`
}

type Base struct {
	Ref string `json:"ref"`
}

type PullRequest struct {
	Merged   bool     `json:"merged"`
	MergedAt string   `json:"merged_at"`
	MergedBy MergedBy `json:"merged_by"`
	Head     Head     `json:"head"`
	Base     Base     `json:"base"`
}

func checkSecretKeyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		mySecret := os.Getenv("SECRET_KEY")

		secreteKeyHeader := c.Request().Header.Get("secretkey")

		if secreteKeyHeader != mySecret {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "unauthorized",
			})
		}

		return next(c)

	}
}

func checkMAC(message []byte, messageMAC, key string) bool {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(message)
	expectedMAC := mac.Sum(nil)

	return hmac.Equal([]byte(messageMAC), []byte(hex.EncodeToString(expectedMAC)))
}

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/static", "static")
	e.Static("/css", "css")
	e.GET("/sse", s.sseHandler)
	e.GET("/", s.HelloWorldHandler)
	e.GET("/health", s.healthHandler)
	e.GET("/updates", s.GetUpdatesHandler)
	e.POST("/github-webhook", s.GithubWebhookHandler)
	e.GET("/home", s.HomeHandler)
	e.POST("/test_sse", s.TestSSEHandler)

	apiGroup := e.Group("/api")
	serverGroup := apiGroup.Group("/server")
	pipelineGroup := apiGroup.Group("/pipeline")

	apiGroup.Use(checkSecretKeyMiddleware)

	serverGroup.POST("/create", s.CreateServerHandler)
	serverGroup.PUT("/update/:id", s.UpdateServerHandler)
	serverGroup.DELETE("/delete/:id", s.DeleteServerHandler)
	serverGroup.GET("/list", s.ListServersHandler)
	serverGroup.GET("/list/:id", s.ListServersHandler)

	pipelineGroup.POST("/create", s.CreatePipelineHandler)
	pipelineGroup.PUT("/update/:id", s.UpdatePipelineHandler)
	pipelineGroup.DELETE("/delete/:id", s.DeletePipelineHandler)
	pipelineGroup.GET("/list", s.ListPipelinesHandler)
	pipelineGroup.POST("/run/:id", s.UpdateProdPipelineHandler)
	pipelineGroup.GET("/check", s.CheckServers)

	e.POST("/create_server", s.CreateServerHandler, checkSecretKeyMiddleware)
	e.PUT("/update_server/:id", s.UpdateServerHandler, checkSecretKeyMiddleware)
	e.DELETE("/delete_server/:id", s.DeleteServerHandler, checkSecretKeyMiddleware)
	e.GET("/list_servers", s.ListServersHandler, checkSecretKeyMiddleware)
	e.POST("/create_pipeline", s.CreatePipelineHandler, checkSecretKeyMiddleware)
	e.PUT("/update_pipeline/:id", s.UpdatePipelineHandler, checkSecretKeyMiddleware)
	e.DELETE("/delete_pipeline/:id", s.DeletePipelineHandler, checkSecretKeyMiddleware)
	e.GET("/list_pipelines", s.ListPipelinesHandler, checkSecretKeyMiddleware)
	e.POST("/update_prod_pipeline/:id", s.UpdateProdPipelineHandler, checkSecretKeyMiddleware)
	e.GET("/check_servers", s.CheckServers, checkSecretKeyMiddleware)
	e.POST("/update_production/:id", s.UpdateProductionById, checkSecretKeyMiddleware)

	return e
}

func (s *Server) TestSSEHandler(c echo.Context) error {
	s.hub.Broadcast <- "update"
	return c.JSON(http.StatusOK, map[string]string{
		"message": "update",
	})
}

func (s *Server) sseHandler(c echo.Context) error {
	client := sse.NewClient(c.Response().Writer)
	s.hub.AddClient <- client
	client.RunSSE()

	return nil
}

func (s *Server) HomeHandler(c echo.Context) error {
	fmt.Println("Home handler called")
	rows, err := s.db.GetUpdates(100, 0)

	if err != nil {
		slog.Error("Error getting updates")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error getting updates",
		})
	}

	return utils.Render(c, views.Index(rows))
}

func (s *Server) GetUpdatesHandler(c echo.Context) error {
	limit, err := strconv.Atoi(c.QueryParam("limit"))

	if err != nil {
		limit = 10
	}
	page, err := strconv.Atoi(c.QueryParam("page"))
	var offset int

	if err != nil {
		offset = 0
	}

	offset = (page - 1) * limit

	updates, err := s.db.GetUpdates(limit, offset)

	if err != nil {
		slog.Error("Error getting updates")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error getting updates",
		})
	}

	return c.JSON(http.StatusOK, updates)

}

func (s *Server) GithubWebhookHandler(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	fmt.Println("body", string(body))

	if err != nil {
		slog.Error("Error reading body")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error reading body",
		})
	}

	for name, headers := range c.Request().Header {
		for _, h := range headers {
			fmt.Printf("%v: %v\n", name, h)
		}
	}

	mySecret := os.Getenv("SECRET_KEY")
	hashSecret := strings.Split(c.Request().Header.Get("X-Hub-Signature-256"), "=")[1]

	if !checkMAC(body, hashSecret, mySecret) {
		slog.Error("Invalid secret")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "invalid secret",
		})
	}

	fmt.Println("secret valid")

	webhook := new(GithubWebhook)
	if err := json.Unmarshal(body, webhook); err != nil {
		slog.Error("Error unmarshalling body")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error unmarshalling body",
		})
	}

	pushDirectlyToDevelop := webhook.Ref == "refs/heads/dev" && !strings.Contains(webhook.HeadCommit.Message, "Merge pull request #")
	pushDirectlyToStaging := webhook.Ref == "refs/heads/staging" && !strings.Contains(webhook.HeadCommit.Message, "Merge pull request #")
	pushDirectlyToMaster := webhook.Ref == "refs/heads/master" && !strings.Contains(webhook.HeadCommit.Message, "Merge pull request #")
	pullRequestMerged := webhook.Action == "closed" && webhook.PullRequest.Merged && webhook.PullRequest.Base.Ref == "master"
	pullRequestMergedToStaging := webhook.Action == "closed" && webhook.PullRequest.Merged && webhook.PullRequest.Base.Ref == "staging"
	pullRequestMergedToDevelop := webhook.Action == "closed" && webhook.PullRequest.Merged && webhook.PullRequest.Base.Ref == "dev"

	fmt.Println("Queue size:", s.queue.Size())

	switch {
	case pushDirectlyToDevelop:
		fmt.Println("pusher name", webhook.Pusher.Name)
		fmt.Println("pusher email", webhook.Pusher.Email)
		fmt.Println("pusher head commit id", webhook.HeadCommit.Id)
		fmt.Println("pusher head commit message", webhook.HeadCommit.Message)

		id, err := s.db.CreateUpdate(webhook.Pusher.Name, "dev", "pending", "in queue")
		// s.hub.Broadcast <- "update"

		if err != nil {
			slog.Error("Error creating update in database")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "error creating update in database",
			})

		}

		options := sshclient.UpdateOptions{
			ID:         id,
			Repository: "dev",
		}

		s.queue.Enqueue(&options)

	case pushDirectlyToStaging:
		fmt.Println("pusher name", webhook.Pusher.Name)
		fmt.Println("pusher email", webhook.Pusher.Email)
		fmt.Println("pusher head commit id", webhook.HeadCommit.Id)
		fmt.Println("pusher head commit message", webhook.HeadCommit.Message)

		id, err := s.db.CreateUpdate(webhook.Pusher.Name, "staging", "pending", "in queue")
		// s.hub.Broadcast <- "update"

		if err != nil {
			slog.Error("Error creating update in database")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "error creating update in database",
			})
		}

		options := sshclient.UpdateOptions{
			ID:         id,
			Repository: "staging",
		}

		s.queue.Enqueue(&options)
	case pushDirectlyToMaster:
		fmt.Println("pusher name", webhook.Pusher.Name)
		fmt.Println("pusher email", webhook.Pusher.Email)
		fmt.Println("pusher head commit id", webhook.HeadCommit.Id)
		fmt.Println("pusher head commit message", webhook.HeadCommit.Message)

		//id, err := s.db.CreateUpdate(webhook.Pusher.Name, "master", "pending", "in queue")
		// s.hub.Broadcast <- "update"

		if err != nil {
			slog.Error("Error creating update in database")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "error creating update in database",
			})
		}

	//	s.queue.Enqueue(id)

	case pullRequestMerged:
		fmt.Println("pull merged", webhook.PullRequest.Merged)
		fmt.Println("pull merged at", webhook.PullRequest.MergedAt)
		fmt.Println("pull merged by", webhook.PullRequest.MergedBy.Login)
		fmt.Println("pull head ref", webhook.PullRequest.Head.Ref)
		fmt.Println("pull head repo", webhook.PullRequest.Head.Repo.FullName)
		fmt.Println("pull base ref", webhook.PullRequest.Base.Ref)

		//id, err := s.db.CreateUpdate(webhook.PullRequest.MergedBy.Login, webhook.PullRequest.Head.Ref, "pending", "in queue")

		// s.hub.Broadcast <- "update"

		if err != nil {
			slog.Error("Error creating update in database")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "error creating update in database",
			})

		}

		// s.queue.Enqueue(id)

	case pullRequestMergedToStaging:
		fmt.Println("pull merged", webhook.PullRequest.Merged)
		fmt.Println("pull merged at", webhook.PullRequest.MergedAt)
		fmt.Println("pull merged by", webhook.PullRequest.MergedBy.Login)
		fmt.Println("pull head ref", webhook.PullRequest.Head.Ref)
		fmt.Println("pull head repo", webhook.PullRequest.Head.Repo.FullName)
		fmt.Println("pull base ref", webhook.PullRequest.Base.Ref)

		id, err := s.db.CreateUpdate(webhook.PullRequest.MergedBy.Login, webhook.PullRequest.Head.Ref, "pending", "in queue")

		// s.hub.Broadcast <- "update"

		if err != nil {
			slog.Error("Error creating update in database")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "error creating update in database",
			})

		}

		options := sshclient.UpdateOptions{
			ID:         id,
			Repository: "staging",
		}

		s.queue.Enqueue(&options)
	case pullRequestMergedToDevelop:
		fmt.Println("pull merged", webhook.PullRequest.Merged)
		fmt.Println("pull merged at", webhook.PullRequest.MergedAt)
		fmt.Println("pull merged by", webhook.PullRequest.MergedBy.Login)
		fmt.Println("pull head ref", webhook.PullRequest.Head.Ref)
		fmt.Println("pull head repo", webhook.PullRequest.Head.Repo.FullName)
		fmt.Println("pull base ref", webhook.PullRequest.Base.Ref)

		id, err := s.db.CreateUpdate(webhook.PullRequest.MergedBy.Login, webhook.PullRequest.Head.Ref, "pending", "in queue")

		// s.hub.Broadcast <- "update"

		if err != nil {
			slog.Error("Error creating update in database")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "error creating update in database",
			})

		}

		options := sshclient.UpdateOptions{
			ID:         id,
			Repository: "dev",
		}

		s.queue.Enqueue(&options)
	}

	slog.Info("Repository added in queue")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "ok",
	})

}

func (s *Server) CreateServerHandler(c echo.Context) error {
	fmt.Println("Criando servidor")

	serverinfo := new(ServerInfo)

	if err := c.Bind(serverinfo); err != nil {
		fmt.Println("error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid request",
		})
	}

	fmt.Println("serverinfo", serverinfo.Host)
	fmt.Println("serverinfo", serverinfo.Password)
	fmt.Println("serverinfo", serverinfo.Script)
	fmt.Println("serverinfo", serverinfo.PipelineID)
	fmt.Println("serverinfo", serverinfo)

	newId, err := s.db.CreateServer(serverinfo.Host, serverinfo.Password, serverinfo.Script, serverinfo.PipelineID, serverinfo.Label)

	if err != nil {
		fmt.Println("error", err)
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

func (s *Server) CreatePipelineHandler(c echo.Context) error {
	name := c.FormValue("name")

	id, err := s.db.CreatePipeline(name)

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

	err = s.db.UpdatePipeline(updatePipeline)

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
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid id",
		})
	}

	err = s.db.DeletePipeline(id)

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
	pipelines, err := s.db.ListPipelines()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error getting pipelines",
		})
	}

	return c.JSON(http.StatusOK, pipelines)

}

func (s *Server) UpdateProdPipelineHandler(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "invalid id",
		})
	}

	go func() {
		sshclient.UpdateProductionNew(id)
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
		sshclient.UpdateProductionById(id)
	}()

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Atualização de produção iniciada com sucesso",
	})

}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) CheckServers(c echo.Context) error {

	website := "https://web.topzap.com.br"

	// check if the website is up

	resp, err := http.Get(website)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "error checking website",
		})
	}

	// print all the response

	fmt.Println("response", resp)

	if resp.StatusCode != 200 {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "website is down",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "website is up",
	})
}
