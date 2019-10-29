package main

import (
	"errors"
	"github.com/jessevdk/go-flags"
	"strings"
)

type Options struct {
	Address     string `short:"l" long:"listen" env:"LISTEN_IP" default:"0.0.0.0" description:"IP address to listen on"`
	Port        int    `short:"p" long:"port" env:"LISTEN_PORT" default:"8080" description:"TCP port number"`
	ShowVersion bool   `short:"V" long:"version" description:"Show version and exit"`
}

// Parse options using fresh parser instance.
func (o *Options) Parse() (*flags.Parser, []string, error) {
	var p = flags.NewParser(o, flags.Default)
	var s, err = p.Parse()

	return p, s, err
}

// Make options check.
func (o *Options) Check() (bool, error) {
	// Check API key
	if len(strings.TrimSpace(o.Address)) < 7 {
		return false, errors.New("wrong address to listen on")
	}

	// Check threads count
	if o.Port <= 0 || o.Port > 65535 {
		return false, errors.New("wrong port number")
	}

	return true, nil
}
