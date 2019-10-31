package main

import (
	"errors"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"regexp"
	"strings"
)

type Options struct {
	Address     string `short:"l" long:"listen" env:"LISTEN_ADDR" default:"0.0.0.0" description:"Address (IP) to listen on"`
	Port        int    `short:"p" long:"port" env:"LISTEN_PORT" default:"8080" description:"TCP port number"`
	ProxyPrefix string `short:"x" long:"prefix" env:"PROXY_PREFIX" default:"proxy" description:"Proxy route prefix"`
	TslCertFile string `long:"tsl-cert" env:"TSL_CERT" description:"TSL certificate file path"`
	TslKeyFile  string `long:"tsl-key" env:"TSL_KEY" description:"TSL key file path"`
	ShowVersion bool   `short:"V" long:"version" description:"Show version and exit"`
	stdLog      *log.Logger
	errLog      *log.Logger
	onExit      OptionsExitFunc
	parseFlags  flags.Options
}

type OptionsExitFunc func(code int)

// Create new options instance.
func NewOptions(stdOut, stdErr *log.Logger, onExit OptionsExitFunc) *Options {
	if onExit == nil {
		onExit = func(code int) {
			os.Exit(code)
		}
	}
	return &Options{
		stdLog:     stdOut,
		errLog:     stdErr,
		onExit:     onExit,
		parseFlags: flags.Default,
	}
}

// Parse options using fresh parser instance.
func (o *Options) Parse() *flags.Parser {
	var parser = flags.NewParser(o, o.parseFlags)
	var _, err = parser.Parse()

	// Parse passed options
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			o.onExit(0)
		} else {
			parser.WriteHelp(o.stdLog.Writer())
			o.onExit(1)
		}
	}

	// Show application version and exit, if flag `-V` passed
	if o.ShowVersion == true {
		o.stdLog.Println("Version: " + VERSION)
		o.onExit(0)
	}

	// Make options check
	if _, err := o.Check(); err != nil {
		o.errLog.Println(err.Error())
		o.onExit(1)
	}

	return parser
}

// Make options check.
func (o *Options) Check() (bool, error) {
	// Check prefix
	if len(strings.TrimSpace(o.ProxyPrefix)) <= 0 || !regexp.MustCompile(`^[a-zA-Z0-9_\-/]+$`).MatchString(o.ProxyPrefix) {
		return false, errors.New("wrong prefix passed")
	}

	// Check address
	if len(strings.TrimSpace(o.Address)) < 7 {
		return false, errors.New("wrong address to listen on")
	}

	// Check port
	if o.Port <= 0 || o.Port > 65535 {
		return false, errors.New("wrong port number")
	}

	// Check TSL cert file path
	if o.TslCertFile != "" {
		if info, err := os.Stat(o.TslCertFile); err != nil || !info.Mode().IsRegular() {
			return false, errors.New("wrong TSL certificate file path")
		}
	}

	// Check TSL key file path
	if o.TslKeyFile != "" {
		if info, err := os.Stat(o.TslKeyFile); err != nil || !info.Mode().IsRegular() {
			return false, errors.New("wrong TSL key file path")
		}
	}

	return true, nil
}
