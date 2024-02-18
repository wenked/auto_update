package sshclient

import (
	"auto-update/internal/database"
	"auto-update/internal/sse"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

type ServerInfo struct {
	Host     string
	Password string
	Folder   string
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

func UpdateRepository(id int64) error {
	fmt.Println("testee2")
	fmt.Println("Atualizando repositório no servidor update com id: ", id)
	auth := goph.Password(os.Getenv("SSH_PASSWORD"))

	client, err := goph.NewConn(&goph.Config{
		User:     "root",
		Addr:     os.Getenv("SSH_HOST"),
		Port:     22,
		Auth:     auth,
		Callback: verifyHost,
	})

	if err != nil {
		fmt.Println("error ao conectar com o servidor", err)
		slog.Error("error ao conectar com o servidor", err)
		return err
	}

	defer client.Close()

	fmt.Println("Conectado com sucesso ao servidor")

	err = database.GetService().UpdateStatusAndMessage(id, "running", "Atualizando repositório no servidor update")
	// sse.GetHub().Broadcast <- "Atualizando repositório no servidor update"

	if err != nil {
		fmt.Println("error ao atualizar status do update", err)
		slog.Error("error ao atualizar status do update", err)
	}

	// out, err := client.Run("cd /topzap-dev/web-greenchat && ls -a")

	out, err := client.Run("cd /topzap-dev/web-greenchat && ls -a && wget -qO- https://raw.githubusercontent.com/nvm-sh/nvm/v0.34.0/install.sh | bash && export NVM_DIR=~/.nvm && source ~/.nvm/nvm.sh && nvm use && git pull && npm install && npm run build")
	// out, err := client.Run("cd " + directory + " && git pull &&  docker-compose up -d --force-recreate --build")

	message := string(out)
	fmt.Println(message)

	slicedMessage := message[len(message)-300:]

	if err != nil {
		err = database.GetService().UpdateStatusAndMessage(id, "error", slicedMessage)

		sse.GetHub().Broadcast <- "error ao executar comando de Atualizar"
		if err != nil {
			fmt.Println("error ao atualizar status do update", err)
			slog.Error("error ao atualizar status do update", err)
		}

		fmt.Println("error ao executar comando de Atualizar", err)
		return err
	}

	err = database.GetService().UpdateStatusAndMessage(id, "success", slicedMessage)

	if err != nil {
		fmt.Println("error ao atualizar status do update", err)
		slog.Error("error ao atualizar status do update", err)
	}

	// sse.GetHub().Broadcast <- "Atualização realizada com sucesso"
	return nil
}

func UpdateProductionNew(pipeline_id int64) error {
	fmt.Println("Atualizando repositório no servidor de produção")

	slog.Info("Atualizando repositório no servidor de produção")

	servers, err := database.GetService().ListServers(pipeline_id)

	if err != nil {
		fmt.Println("error ao buscar servidores", err)
		slog.Error("error ao buscar servidores", err)
		return err
	}

	// wait group
	var wg sync.WaitGroup

	for _, server := range servers {
		wg.Add(1)
		go func(server database.UpdateServer) {
			defer wg.Done()
			fmt.Println("Atualizando repositório no servidor de produção", server.Host)

			auth := goph.Password(server.Password)

			client, err := goph.NewConn(&goph.Config{
				User:     "root",
				Addr:     server.Host,
				Port:     22,
				Auth:     auth,
				Callback: verifyHost,
			})

			if err != nil {
				fmt.Println("error ao conectar com o servidor:"+server.Host, err)
				slog.Error("error ao conectar com o servidor", err)
				return
			}

			defer client.Close()

			// run script

			out, err := client.Run(server.Script)

			message := string(out)

			fmt.Println(message)

			if err != nil {
				fmt.Println("error ao executar comando de Atualizar o servidor:"+server.Host, err)
				slog.Error("error ao executar comando de Atualizar o servidor", err)
				return
			}

			fmt.Println("Atualização realizada com sucesso:", server.Host)

		}(server)
	}

	wg.Wait()
	return nil
}
