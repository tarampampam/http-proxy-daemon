package cmd

import (
	"http-proxy-daemon/cmd/serve"
	"http-proxy-daemon/cmd/version"
)

// Root is a basic commands struct.
type Root struct {
	Version version.Command `command:"version" alias:"v" description:"Display application version"`
	Serve   serve.Command   `command:"serve" alias:"s" description:"Start proxy server"`
}
