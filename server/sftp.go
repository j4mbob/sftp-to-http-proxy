package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sftp-to-http-proxy/loader"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func StartServer(cliArgs *loader.Args) {

	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {

			if conn.User() == cliArgs.UserName && string(password) == cliArgs.Password {
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

	keyBytes, err := os.ReadFile(cliArgs.SSLKey)
	if err != nil {
		log.Fatal("failed to load private key", err)
	}

	key, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		log.Fatal("failed to parse private key", err)
	}

	config.AddHostKey(key)

	listener, err := net.Listen("tcp", cliArgs.ListenIP+":"+cliArgs.ListenPort)
	if err != nil {
		log.Printf("failed to bind server: %v", err)
		os.Exit(1)
	} else {
		log.Printf("sftp proxy listening on %s:%s", cliArgs.ListenIP, cliArgs.ListenPort)

	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept incoming connection: %v\n", err)
			continue
		}
		go handleConn(conn, config, cliArgs)
	}
}

func handleConn(conn net.Conn, config *ssh.ServerConfig, cliArgs *loader.Args) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Printf("failed to handshake: %v\n", err)
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("could not accept channel: %v\n", err)
			continue
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				if req.Type == "subsystem" && string(req.Payload[4:]) == "sftp" {
					handlers := sftp.Handlers{
						FileGet:  customFileReader{remoteUrl: cliArgs.RemoteURL},
						FilePut:  customFileWriter{},
						FileCmd:  customFileCmder{},
						FileList: customFileLister{},
					}

					server := sftp.NewRequestServer(channel, handlers)

					if err := server.Serve(); err == io.EOF {
						server.Close()
						return
					} else if err != nil {
						log.Printf("SFTP server closed with error: %v\n", err)
					}
					return
				}
			}
		}(requests)
	}
}
