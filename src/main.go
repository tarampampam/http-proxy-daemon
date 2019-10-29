package main

import (
	"log"
	"os"
)

const VERSION = "0.0.1" // Do not forget update this value before new version releasing

func main() {
	options := NewOptions(os.Stdout, os.Stderr, func(code int) {
		os.Exit(code)
	})

	options.Parse()

	srv := NewServer(
		options.Address,
		options.Port,
		options.ProxyPrefix,
		log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds),
	)

	srv.RegisterHandlers()

	if err := srv.Start(); err != nil {
		panic(err)
	}
}
