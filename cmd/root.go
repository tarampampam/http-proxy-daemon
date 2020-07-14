package cmd

import "http-proxy-daemon/cmd/version"

// Root is a basic commands struct.
type Root struct {
	Version version.Command `command:"version" alias:"v" description:"Display application version"`
}
