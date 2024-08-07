package sshclient

import (
	"auto-update/internal/database"
	"auto-update/internal/database/models"
	notification "auto-update/internal/notifications"
	"auto-update/internal/sse"
	"auto-update/utils"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

type SshClient interface {
	UpdateRepository(options *UpdateOptions) error
	UpdateProductionNew(pipeline_id int64, userId int64) error
	UpdateProductionById(id int64) error
}

type SshClientService struct {
	db database.Service
}

type ServerInfo struct {
	Host     string
	Password string
	Folder   string
}

type UpdateOptions struct {
	ID         int64
	Repository string
}

type ErrorMessage struct {
	Label  string
	Reason string
}

func NewSshClientService() *SshClientService {
	return &SshClientService{
		db: database.GetService(),
	}
}

func verifyHost(host string, remote net.Addr, key ssh.PublicKey) error {

	hostFound, err := goph.CheckKnownHost(host, remote, key, "")

	if hostFound && err != nil {
		return err
	}

	// handshake because public key already exists.
	if hostFound && err == nil {
		return nil
	}

	return goph.AddKnownHost(host, remote, key, "")
}

func (s *SshClientService) UpdateRepository(options *UpdateOptions) error {
	fmt.Println("Atualizando repositório no servidor update com id: ", options.ID)
	auth := goph.Password(os.Getenv("SSH_PASSWORD"))

	client, err := goph.NewConn(&goph.Config{
		User:     "root",
		Addr:     os.Getenv("SSH_HOST"),
		Port:     22,
		Auth:     auth,
		Callback: verifyHost,
	})

	if err != nil {
		slog.Error("error ao conectar com o servidor", "error", err)
		return err
	}

	defer client.Close()

	fmt.Println("Conectado com sucesso ao servidor")

	err = database.GetService().UpdateStatusAndMessage(options.ID, "running", "Atualizando repositório no servidor update")
	// sse.GetHub().Broadcast <- "Atualizando repositório no servidor update"

	if err != nil {
		slog.Error("error ao atualizar status do update", "error", err)
	}

	var folder string
	runScript := ""
	switch options.Repository {
	case "dev":
		folder = "topzap-dev"
		runScript = fmt.Sprintf("cd /%s/web-greenchat && git pull && docker-compose -f docker-compose-staging.yml up -d --force-recreate --build", folder)

	case "staging":
		folder = "topzap"
		runScript = fmt.Sprintf("cd /%s/web-greenchat && ls -a && wget -qO- https://raw.githubusercontent.com/nvm-sh/nvm/v0.34.0/install.sh | bash && export NVM_DIR=~/.nvm && source ~/.nvm/nvm.sh && nvm use &&  pm2 stop all && git pull && npm install && npm run build && pm2 start all", folder)

	default:
		folder = "topzap-dev"
	}

	out, err := client.Run(runScript)

	message := string(out)
	fmt.Println(message)

	slicedMessage := message[len(message)-300:]

	if err != nil {
		err = database.GetService().UpdateStatusAndMessage(options.ID, "error", slicedMessage)

		sse.GetHub().Broadcast <- "error ao executar comando de Atualizar"
		if err != nil {
			fmt.Println("error ao atualizar status do update", err)
			slog.Error("error ao atualizar status do update", err)
		}

		fmt.Println("error ao executar comando de Atualizar", err)
		return err
	}

	err = database.GetService().UpdateStatusAndMessage(options.ID, "success", slicedMessage)

	if err != nil {
		slog.Error("error ao atualizar status do update", "error", err)
	}

	// sse.GetHub().Broadcast <- "Atualização realizada com sucesso"
	return nil
}

func (s *SshClientService) UpdateProductionNew(pipeline_id int64, userId int64) error {
	slog.Info("Atualizando repositório no servidor de produção")
	db := database.GetService()

	pipeline, err := db.GetUserPipelineById(pipeline_id, userId)

	if err != nil {
		slog.Error("error finding pipeline", "error", err)
		return err
	}

	servers, err := db.ListServers(pipeline_id)

	notificationService := notification.NewNotificationService()

	if err != nil {
		slog.Error("error ao buscar servidores", "error", err)
		return err
	}

	err = notificationService.SendAllNotifications(fmt.Sprintf("Atualização iniciada na pipeline: *%s*", pipeline.Name), userId, "yellow")

	if err != nil {
		slog.Error("error ao enviar notificação", "error", err)
		return err
	}

	errors := make([]ErrorMessage, 0)

	var wg sync.WaitGroup

	for _, server := range servers {
		wg.Add(1)
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)

		go func(ctx context.Context, server models.UpdateServer) {

			defer wg.Done()
			defer cancel()

			done := make(chan bool)

			go func() {
				slog.Info("Atualizando repositório no servidor de produção", server.Label)

				decryptedPassword, err := utils.Decrypt(server.Password)

				if err != nil {
					slog.Error("error ao decriptar password com o servidor", "error", err)

					errors = append(errors, ErrorMessage{Label: server.Label, Reason: err.Error()})
					done <- true
					return
				}

				auth := goph.Password(decryptedPassword)

				client, err := goph.NewConn(&goph.Config{
					User:     "root",
					Addr:     server.Host,
					Port:     22,
					Auth:     auth,
					Callback: verifyHost,
				})

				if err != nil {
					fmt.Println("error ao conectar com o servidor:"+server.Host, err)
					slog.Error("error ao conectar com o servidor", "error", err)

					errors = append(errors, ErrorMessage{Label: server.Label, Reason: err.Error()})
					done <- true
					return
				}

				defer client.Close()

				// run script

				out, err := client.Run(server.Script)
				//out, err := client.Run("ls -a")

				message := string(out)

				fmt.Println(message)

				if err != nil {
					slog.Error("error ao executar comando de Atualizar o servidor:"+server.Host, "error", err)

					errors = append(errors, ErrorMessage{Label: server.Label, Reason: message})
				}

				done <- true
			}()

			select {
			case <-ctx.Done():
				slog.Info("Timeout reached for server", "info", server.Label)
			case <-done:
				slog.Info("Atualização realizada com sucesso:", "info", server.Label)
			}

		}(ctx, server)
	}

	wg.Wait()

	var msg strings.Builder
	color := "green"
	msg.WriteString(fmt.Sprintf("Atualização realizada com sucesso na pipeline: *%s*", pipeline.Name))

	if len(errors) > 0 {
		msg.WriteString("\n\nErros encontrados nos servidores:\n")
		for _, e := range errors {
			msg.WriteString(fmt.Sprintf("```*%s* - %s```\n", e.Label, e.Reason))
		}
		color = "red"
	}

	fmt.Println(msg.String())
	err = notificationService.SendAllNotifications(msg.String(), userId, color)

	if err != nil {
		slog.Error("error ao enviar notificação", "error", err)
		return err
	}
	return nil
}

func (s *SshClientService) UpdateProductionById(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()

	slog.Info("Atualizando repositório no servidor de produção")

	server, err := database.GetService().GetServer(id)

	if err != nil {

		slog.Error("error ao buscar servidores", "error", err)
		return err
	}

	done := make(chan bool)

	slog.Info("Atualizando repositório no servidor de produção", "info", server.Label)
	go func() {

		decryptedPassword, err := utils.Decrypt(server.Password)
		if err != nil {
			slog.Error("error ao decriptar o password:"+server.Host, "error", err)
			done <- true
			return
		}

		auth := goph.Password(decryptedPassword)

		client, err := goph.NewConn(&goph.Config{
			User:     "root",
			Addr:     server.Host,
			Port:     22,
			Auth:     auth,
			Callback: verifyHost,
		})

		if err != nil {
			slog.Error("error ao conectar com o servidor:"+server.Host, "error", err)
			done <- true
			return
		}

		defer client.Close()

		// run script

		out, err := client.Run(server.Script)
		//out, err := client.Run("ls -a")

		message := string(out)

		fmt.Println(message)

		if err != nil {
			slog.Error("error ao executar comando de Atualizar o servidor:"+server.Host, "error", err)
		}

		done <- true
	}()

	select {
	case <-ctx.Done():
		fmt.Println("Timeout reached for server", server.Label)
	case <-done:
		fmt.Println("Atualização realizada com sucesso:", server.Label)
	}

	return nil
}
