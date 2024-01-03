package loader

import (
	"fmt"
	"os"
	"runtime"

	"github.com/grafana/pyroscope-go"
)

func Startup() *Args {

	arguments := new(Args)

	argParse(arguments)

	implementPID(arguments.PidFile)

	if arguments.Pyroscope {

		fmt.Printf("sending application metrics to remote pyroscope host: %s", arguments.PyroscopeHost)
		StartPyroScope(arguments)

	}

	return arguments
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
