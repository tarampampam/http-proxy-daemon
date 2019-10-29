package main

import (
	"github.com/jessevdk/go-flags"
	"log"
	"os"
)

const VERSION = "0.0.1" // Do not forget update this value before new version releasing

var options Options

func main() {
	// Parse passed options
	if parser, _, err := options.Parse(); parser != nil && err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			parser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}

	// Show application version and exit, if flag `-V` passed
	if options.ShowVersion == true {
		_, _ = os.Stdout.WriteString("Version: " + VERSION + "\n")
		os.Exit(0)
	}

	// Make options check
	if _, err := options.Check(); err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	srv := NewServer("0.0.0.0", "8080", log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds))

	srv.RegisterHandlers()

	if err := srv.Start(); err != nil {
		panic(err)
	}
}
