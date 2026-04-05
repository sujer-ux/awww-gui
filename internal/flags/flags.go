package flags

import (
	"flag"
	"os"
)

type Flags struct {
	LogLevel string
	Deamon   bool
}

func Get() Flags {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	logLevel := flag.String("log-level", "info", "log level")
	deamon := flag.Bool("start deamon", false, "")

	flag.Parse()

	return Flags{
		LogLevel: *logLevel,
		Deamon:   *deamon,
	}
}
