package serve

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/tarampampam/http-proxy-daemon/internal/pkg/config"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/env"

	"github.com/spf13/pflag"
)

type flags struct {
	listen struct {
		ip   string
		port uint16
	}

	proxy struct {
		routePrefix    string
		requestTimeout time.Duration
	}
}

func (f *flags) init(flagSet *pflag.FlagSet) {
	flagSet.StringVarP(
		&f.listen.ip,
		"listen",
		"l",
		"0.0.0.0",
		fmt.Sprintf("IP address to listen on [$%s]", env.ListenAddr),
	)
	flagSet.Uint16VarP(
		&f.listen.port,
		"port",
		"p",
		8080, //nolint:gomnd
		fmt.Sprintf("TCP port number [$%s]", env.ListenPort),
	)
	flagSet.StringVarP(
		&f.proxy.routePrefix,
		"prefix",
		"x",
		"proxy",
		fmt.Sprintf("Proxy route prefix [$%s]", env.ProxyRoutePrefix),
	)
	flagSet.DurationVarP(
		&f.proxy.requestTimeout,
		"proxy-request-timeout",
		"",
		time.Second*30, //nolint:gomnd
		fmt.Sprintf("Proxy request timeout (examples: 5s, 15s30ms) [$%s]", env.ProxyRequestTimeout),
	)
}

func (f *flags) overrideUsingEnv() error {
	if envVar, exists := env.ListenAddr.Lookup(); exists {
		f.listen.ip = envVar
	}

	if envVar, exists := env.ListenPort.Lookup(); exists {
		if p, err := strconv.ParseUint(envVar, 10, 16); err == nil { //nolint:gomnd
			f.listen.port = uint16(p)
		} else {
			return fmt.Errorf("wrong TCP port environment variable [%s] value", envVar)
		}
	}

	if envVar, exists := env.ProxyRoutePrefix.Lookup(); exists {
		f.proxy.routePrefix = envVar
	}

	if envVar, exists := env.ProxyRequestTimeout.Lookup(); exists {
		if d, err := time.ParseDuration(envVar); err == nil {
			f.proxy.requestTimeout = d
		} else {
			return fmt.Errorf("wrong proxy request timeout [%s] value", envVar)
		}
	}

	return nil
}

func (f *flags) validate() error {
	if net.ParseIP(f.listen.ip) == nil {
		return fmt.Errorf("wrong IP address [%s] for listening", f.listen.ip)
	}

	if f.proxy.routePrefix == "" {
		return errors.New("empty proxy route prefix")
	} else if !regexp.MustCompile(`^[a-zA-Z0-9_\-/]+$`).MatchString(f.proxy.routePrefix) {
		return fmt.Errorf("wrong proxy prefix [%s] value", f.proxy.routePrefix)
	}

	return nil
}

func (f *flags) toConfig() config.Config {
	cfg := config.Config{}

	cfg.Proxy.Prefix = f.proxy.routePrefix
	cfg.Proxy.RequestTimeout = f.proxy.requestTimeout

	return cfg
}
