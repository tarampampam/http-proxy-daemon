package config

import "time"

// Config is application runtime configuration.
type Config struct {
	Proxy struct {
		Prefix         string
		RequestTimeout time.Duration
	}
}
