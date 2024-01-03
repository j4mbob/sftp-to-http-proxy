package server

import (
	"fmt"
	"net"
	"os"
	"sftp-to-http-proxy/loader"

	"golang.org/x/crypto/ssh"
)

func StartProxy(cliArgs *loader.Args) {
	sshConfig := setupServer(cliArgs.UserName, cliArgs.Password)

	loadKey(sshConfig, cliArgs.SSLKey, cliArgs.PidFile)

	listener := startListener(cliArgs.ListenIP, cliArgs.ListenPort, cliArgs.PidFile)

	acceptConnections(listener, sshConfig, cliArgs.RemoteURL)

}

func setupServer(userName string, password string) *ssh.ServerConfig {
	sshConfig := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, enteredPassword []byte) (*ssh.Permissions, error) {

			if conn.User() == userName && string(enteredPassword) == password {
				fmt.Printf("successful login from: %s", conn.RemoteAddr().String())

				return nil, nil
			} else {
				fmt.Printf("failed login from: %s", conn.RemoteAddr().String())
			}
			errorMsg := fmt.Sprintf("password rejected for user: %s from address: %s", conn.User(), conn.RemoteAddr().String())
			return nil, fmt.Errorf(errorMsg)
		},
		NoClientAuth: false,
	}

	return sshConfig
}

func loadKey(config *ssh.ServerConfig, sslKey string, pidFile string) {
	keyBytes, err := os.ReadFile(sslKey)
	if err != nil {
		fmt.Printf("failed to load private key: %v", err)
		loader.CleanUp(pidFile)
	}

	key, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		fmt.Printf("failed to parse private key: %v", err)
		loader.CleanUp(pidFile)
	}

	config.AddHostKey(key)

}

func startListener(listenIP string, listenPort string, pidFile string) net.Listener {
	listener, err := net.Listen("tcp", listenIP+":"+listenPort)
	if err != nil {
		fmt.Printf("failed to bind server: %v", err)
		os.Exit(1)
	} else {
		fmt.Printf("sftp proxy listening on %s:%s", listenIP, listenPort)

	}

	return listener

}

func acceptConnections(listener net.Listener, sshConfig *ssh.ServerConfig, remoteUrl string) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("failed to accept incoming connection: %v\n", err)
			continue
		}
		go newConnection(conn, sshConfig, remoteUrl)
	}

}
