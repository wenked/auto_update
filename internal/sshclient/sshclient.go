package sshclient

import (
	"fmt"
	"net"
	"os"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

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

func UpdateRepository() (error) {
	auth := goph.Password(os.Getenv("SSH_PASSWORD"))

	client, err := goph.NewConn(&goph.Config{
		User:     "root",
		Addr:    os.Getenv("SSH_HOST"),
		Port:     22,
		Auth:     auth,
		Callback: verifyHost,
	})


	if err != nil {
		fmt.Println("error ao conectar com o servidor")
		return err
	}

	defer client.Close()


	fmt.Println("Conectado com sucesso ao servidor")

	out, err := client.Run("cd /topzap-dev/web-greenchat && ls -a && wget -qO- https://raw.githubusercontent.com/nvm-sh/nvm/v0.34.0/install.sh | bash && export NVM_DIR=~/.nvm && source ~/.nvm/nvm.sh && nvm use && git pull && npm install && npm run build")
	// out, err := client.Run("cd " + directory + " && git pull &&  docker-compose up -d --force-recreate --build")

	fmt.Println(string(out))


	if err != nil {
		fmt.Println("error ao executar comando de Atualizar",err)
		return err
	}

	return nil
}