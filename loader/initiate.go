package loader

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"syscall"

	"github.com/grafana/pyroscope-go"
)

func Startup() *Args {
	const pidFile = "/run/sftp-proxy.pid"

	implementPID(pidFile)

	arguments := new(Args)

	argParse(arguments)

	if arguments.Pyroscope {

		fmt.Printf("sending application metrics to remote pyroscope host: %s", arguments.PyroscopeHost)
		StartPyroScope(arguments)

	}

	return arguments
}

func implementPID(pidFile string) {
	if checkPID(pidFile) {
		fmt.Printf("Another instance of sftp-proxy is already running. Exiting.")
		os.Exit(1)
	}

	err := writePID(pidFile)
	if err != nil {
		fmt.Printf("Unable to write PID file: %s", err)
		os.Exit(1)
	}

}

func writePID(pidFile string) error {
	pid := []byte(strconv.Itoa(os.Getpid()) + "\n")
	return os.WriteFile(pidFile, pid, 0644)
}

func checkPID(pidFile string) bool {
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		fmt.Printf("Invalid PID in PID file: %s", pidData)
		os.Exit(1)
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func StartPyroScope(arguments *Args) {
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	pyroscope.Start(pyroscope.Config{
		ApplicationName: "sftp-to-http-proxy", ServerAddress: arguments.PyroscopeHost, Logger: nil,
		Tags: map[string]string{"hostname": os.Getenv("HOSTNAME"), "application": "sftp-to-http-proxy"},
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})

}
