package flags

import (
	"flag"
	"os"
)

type Flags struct {
	LogLevel string
	Daemon   bool
}

func Get() Flags {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	logLevel := flag.String("log-level", "info", "log level")
	daemon := flag.Bool("d", false, "")

	flag.Parse()

	return Flags{
		LogLevel: *logLevel,
		Daemon:   *daemon,
	}
}
