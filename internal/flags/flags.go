package flags

import (
	"flag"
	"os"
)

type Flags struct {
	LogLevel string
	Daemon   bool
	Kill     bool
	Random   bool
}

func Get() Flags {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	logLevel := flag.String("log-level", "info", "log level")
	daemon := flag.Bool("d", false, "Run as daemon mode")
	kill := flag.Bool("kill", false, "Kill daemon")
	random := flag.Bool("random", false, "Set random wallpaper")

	flag.Parse()

	return Flags{
		LogLevel: *logLevel,
		Daemon:   *daemon,
		Kill:     *kill,
		Random:   *random,
	}
}
