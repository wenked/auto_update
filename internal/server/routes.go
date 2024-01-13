package server

import (
	"auto-update/internal/sshclient"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type GithubWebhook struct {
	Ref string `json:"ref"`
	Pusher Pusher `json:"pusher"`
	HeadCommit HeadCommit `json:"head_commit"`
	Action string `json:"action"`
   	PullRequest PullRequest `json:"pull_request"`
}

type HeadCommit struct {
	Id string `json:"id"`
	Message string `json:"message"`
}
type Pusher struct {
	Name string `json:"name"`
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

	e.GET("/", s.HelloWorldHandler)
	e.GET("/health", s.healthHandler)
	e.POST("/github-webhook", s.GithubWebhookHandler)

	return e
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

	webhook := new(GithubWebhook)
    if err := c.Bind(webhook); err != nil {
        return err
    }
	
	pushDirectlyToMaster := webhook.Ref == "refs/heads/master"
	pullRequestMerged := webhook.Action == "closed" && webhook.PullRequest.Merged && webhook.PullRequest.Base.Ref == "master"

	if(pushDirectlyToMaster){
		fmt.Println("pusher name", webhook.Pusher.Name)
		fmt.Println("pusher email", webhook.Pusher.Email)
		fmt.Println("pusher head commit id", webhook.HeadCommit.Id)
		fmt.Println("pusher head commit message", webhook.HeadCommit.Message)

		err := sshclient.UpdateRepository()

		if err != nil {
			slog.Error("Error update repository")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "error update repository",
			})
		}

		slog.Info("Repository updated")
		return c.JSON(http.StatusOK, map[string]string{
			"message": "update repository",
		})
	}

	if(pullRequestMerged){
		fmt.Println("pull merged", webhook.PullRequest.Merged)
		fmt.Println("pull merged at", webhook.PullRequest.MergedAt)
		fmt.Println("pull merged by", webhook.PullRequest.MergedBy.Login)
		fmt.Println("pull head ref", webhook.PullRequest.Head.Ref)
		fmt.Println("pull head repo", webhook.PullRequest.Head.Repo.FullName)
		fmt.Println("pull base ref", webhook.PullRequest.Base.Ref)

		err := sshclient.UpdateRepository()

		if err != nil {
			slog.Error("Error update repository")
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "error update repository",
			})
		}

		slog.Info("Repository updated")
		return c.JSON(http.StatusOK, map[string]string{
			"message": "update repository",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "pull request not merged",
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
