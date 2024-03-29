package server

import (
	"fmt"
	"io"
	"net"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func newConnection(conn net.Conn, sshConfig *ssh.ServerConfig, remoteUrl string) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, sshConfig)
	if err != nil {
		fmt.Printf("failed to handshake: %v\n", err)
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, server, err := handleNewChannel(newChannel, remoteUrl, conn)
		if err != nil {
			fmt.Printf("could not accept channel: %v\n", err)
			continue
		}
		go handleRequests(requests, channel, server, conn)
	}
}

func handleNewChannel(newChannel ssh.NewChannel, remoteUrl string, conn net.Conn) (ssh.Channel, <-chan *ssh.Request, *sftp.RequestServer, error) {
	channel, requests, err := newChannel.Accept()
	if err != nil {
		return nil, nil, nil, err
	}

	handlers := createHandlers(remoteUrl, conn.RemoteAddr().String())
	server := sftp.NewRequestServer(channel, handlers)

	return channel, requests, server, nil
}

func createHandlers(remoteUrl string, clientIP string) sftp.Handlers {
	return sftp.Handlers{
		FileGet:  customFileReader{remoteUrl: remoteUrl, clientIP: clientIP},
		FilePut:  customFileWriter{},
		FileCmd:  customFileCmder{},
		FileList: customFileLister{},
	}
}

func handleRequests(requests <-chan *ssh.Request, channel ssh.Channel, server *sftp.RequestServer, conn net.Conn) {
	for req := range requests {
		if req.Type == "subsystem" && string(req.Payload[4:]) == "sftp" {
			req.Reply(true, nil)
			if err := checkServerState(channel, server, conn); err != nil {
				fmt.Printf("SFTP server closed with error: %v\n", err)
			}
			return
		}
	}
}

func checkServerState(channel ssh.Channel, server *sftp.RequestServer, conn net.Conn) error {
	if err := server.Serve(); err != nil {
		if err == io.EOF {
			fmt.Printf("client %s closed connection\n", conn.RemoteAddr().String())
			server.Close()
			return nil
		}
		return err
	}
	return nil
}
