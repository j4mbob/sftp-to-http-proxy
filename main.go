package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// scrappy qucik poc code for just proving it works; needs some (alot) of refinement to make it production

func main() {

	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {

			if conn.User() == "jamie" && string(password) == "test" {

				return nil, nil
			}

			return nil, fmt.Errorf("password rejected for %q", conn.User())
		},
		NoClientAuth: false,
	}

	privateBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		log.Fatal("Failed to load private key", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key", err)
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		log.Printf("Failed to listen on 2022: %v\n", err)
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection: %v\n", err)
			continue
		}
		go handleConn(conn, config)
	}
}

type customFileWriter struct{}

func (c customFileWriter) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	return nil, nil
}

type customFileCmder struct{}

func (c customFileCmder) Filecmd(r *sftp.Request) error {
	fmt.Println("cmd being called")
	return nil
}

// fake a file as an sftp client needs to get the file stats before it will issue a get
// this works fine in a normal setting when sftp is serving from the local filesystem, but as are we
// injecting a custom get handler which is reading from a remote http server we need to to trick the
// client into issuing the get by feeding it fake file stats

type fakeFile struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (f *fakeFile) Name() string       { return f.name }
func (f *fakeFile) Size() int64        { return f.size }
func (f *fakeFile) Mode() os.FileMode  { return 0666 }
func (f *fakeFile) ModTime() time.Time { return f.modTime }
func (f *fakeFile) IsDir() bool        { return f.isDir }
func (f *fakeFile) Sys() interface{}   { return nil }

type listerat []os.FileInfo

type customFileLister struct{}

func (l listerat) ListAt(p []os.FileInfo, off int64) (int, error) {
	copied := copy(p, l[off:])
	return copied, nil
}

func (c customFileLister) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	log.Printf("DEBUG: method called: %s", r.Method)

	switch r.Method {
	case "List":
		var files listerat = []os.FileInfo{&fakeFile{name: "fakefile.txt"}}
		return files, nil
	case "Stat":
		var file listerat = []os.FileInfo{&fakeFile{name: "fakefile.txt"}}
		return file, nil
	}

	return nil, errors.New("unsupported operation")
}

type customFileReader struct{}

func (c customFileReader) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	log.Print("DEBUG: reader called")

	resp, err := http.Get("http://grafana.networks-util.ask4.net:8080/" + r.Filepath)
	if err != nil {

		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {

		return nil, err
	}

	return bytes.NewReader(data), nil
}

func handleConn(conn net.Conn, config *ssh.ServerConfig) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Printf("Failed to handshake: %v\n", err)
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
			log.Printf("Could not accept channel: %v\n", err)
			continue
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				if req.Type == "subsystem" && string(req.Payload[4:]) == "sftp" {
					handlers := sftp.Handlers{
						FileGet:  customFileReader{},
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
