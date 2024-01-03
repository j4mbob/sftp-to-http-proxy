package loader

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

type Args struct {
	ListenIP      string `json:"listenip"`
	ListenPort    string `json:"listenport"`
	UserName      string `json:"username"`
	Password      string `json:"password"`
	SSLKey        string `json:"sslkey"`
	RemoteURL     string `json:"remoteurl"`
	Pyroscope     bool   `json:"pyroscope"`
	PyroscopeHost string `json:"pyroscopehost"`
	PidFile       string `json:"pidfile"`
}

func argParse(arguments *Args) {
	listenIp := flag.String("listenip", "0.0.0.0", "IP for SFTP server to bind to")
	listenPort := flag.String("listenport", "2122", "port for SFTP server to listen on")
	userName := flag.String("username", "sftp", "username to use for authentication")
	password := flag.String("password", "sftp", "password to use for authentication")
	sslKey := flag.String("sslkey", "id_rsa", "ssl private key to use")
	remoteUrl := flag.String("remoteurl", "http://grafana.networks-util.ask4.net:8080", "remote web server to send requests to")
	pyroscope := flag.Bool("pyroscope", false, "enable sending application metrics to pyroscope host")
	pyroscopeHost := flag.String("pyroscopehost", "http://grafana.networks-util.ask4.net", "remote pyroscope to send application metrics to")
	pidFile := flag.String("pidfile", "/run/sftp-proxy.pid", "PID file location to use")

	configFile := flag.String("loadconfig", "none", "load json config file")

	flag.Parse()

	arguments.ListenIP = *listenIp
	arguments.ListenPort = *listenPort
	arguments.UserName = *userName
	arguments.Password = *password
	arguments.SSLKey = *sslKey
	arguments.RemoteURL = *remoteUrl
	arguments.PyroscopeHost = *pyroscopeHost
	arguments.Pyroscope = *pyroscope
	arguments.PidFile = *pidFile

	if *configFile != "none" {
		fmt.Printf("loading JSON config: %s\n", *configFile)
		loadConfig(*configFile, arguments)
	}

}

func loadConfig(configFile string, arguments *Args) {
	jsonFile, err := os.Open(configFile)

	if err != nil {
		fmt.Println(err)
		CleanUp(arguments.PidFile)

	}

	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), &arguments)

}
