package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

const VERSION = "0.0.1" // Do not forget update this value before new version releasing

func main() {
	var (
		stdLog = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds)
		errLog = log.New(os.Stderr, "", log.LstdFlags)
	)

	// Precess CLI options
	options := NewOptions(stdLog, errLog, func(code int) {
		os.Exit(code)
	})

	// Parse options and make all checks
	options.Parse()

	// Create server instance
	srv := NewServer(
		options.Address,
		options.Port,
		options.ProxyPrefix,
		stdLog,
		errLog,
	)

	// Register server handlers
	srv.RegisterHandlers()

	// Make a channel for system signals
	signals := make(chan os.Signal, 1)

	// "Subscribe" for system signals
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGKILL, syscall.SIGQUIT)

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			errLog.Println(err.Error())
			os.Exit(1)
		}
	}()

	// Listen for a signal
	s := <-signals
	stdLog.Printf("Signal [%v] catched", s)
	if err := srv.Stop(); err != nil {
		errLog.Println(err.Error())
	}
}
