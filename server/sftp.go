package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"sftp-to-http-proxy/loader"

	"golang.org/x/crypto/ssh"
)

func StartProxy(cliArgs *loader.Args) {

	sshConfig := setupServer(cliArgs.UserName, cliArgs.Password)

	loadKey(sshConfig, cliArgs.SSLKey)

	listener := startListener(cliArgs.ListenIP, cliArgs.ListenPort)

	acceptConnections(listener, sshConfig, cliArgs.RemoteURL)

}

func setupServer(userName string, password string) *ssh.ServerConfig {

	sshConfig := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, enteredPassword []byte) (*ssh.Permissions, error) {

			if conn.User() == userName && string(enteredPassword) == password {
				log.Printf("successful login from: %s", conn.RemoteAddr().String())

				return nil, nil
			} else {
				log.Printf("failed login from: %s", conn.RemoteAddr().String())
			}
			errorMsg := fmt.Sprintf("password rejected for user: %s from address: %s", conn.User(), conn.RemoteAddr().String())
			return nil, fmt.Errorf(errorMsg)
		},
		NoClientAuth: false,
	}

	return sshConfig
}

func loadKey(config *ssh.ServerConfig, sslKey string) {

	keyBytes, err := os.ReadFile(sslKey)
	if err != nil {
		log.Fatal("failed to load private key", err)
	}

	key, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		log.Fatal("failed to parse private key", err)
	}

	config.AddHostKey(key)

}

func startListener(listenIP string, listenPort string) net.Listener {

	listener, err := net.Listen("tcp", listenIP+":"+listenPort)
	if err != nil {
		log.Printf("failed to bind server: %v", err)
		os.Exit(1)
	} else {
		log.Printf("sftp proxy listening on %s:%s", listenIP, listenPort)

	}

	return listener

}

func acceptConnections(listener net.Listener, sshConfig *ssh.ServerConfig, remoteUrl string) {

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept incoming connection: %v\n", err)
			continue
		}
		go newConnection(conn, sshConfig, remoteUrl)
	}

}
