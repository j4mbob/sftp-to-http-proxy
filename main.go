package main

import (
	"sftp-to-http-proxy/loader"
	"sftp-to-http-proxy/server"
)

// scrappy qucik poc code for just proving it works; needs some (alot) of refinement to make it production
// TODO:
// PID handling for conversion to system service
// pyroscope client added for application metrics

func main() {

	cliArgs := loader.Startup()
	server.StartServer(cliArgs)
}
