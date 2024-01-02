package main

import (
	"sftp-to-http-proxy/loader"
	"sftp-to-http-proxy/server"
)

func main() {
	cliArgs := loader.Startup()
	server.StartProxy(cliArgs)
}
