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
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	logLevel := flag.String("log-level", "info", "log level")
	daemon := flag.Bool("d", false, "Run as daemon mode")

	fs.Parse(os.Args[1:])

	return Flags{
		LogLevel: *logLevel,
		Daemon:   *daemon,
	}
}
