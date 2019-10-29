package main

import (
	"log"
	"os"
)

func main() {
	srv := NewServer("0.0.0.0", "8080", log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds))

	srv.RegisterHandlers()

	if err := srv.Start(); err != nil {
		panic(err)
	}
}
