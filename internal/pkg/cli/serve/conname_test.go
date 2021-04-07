package serve_test

import (
	"context"
	"net"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/kami-zh/go-capturer"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/cli/serve"
	"go.uber.org/zap"
)

func TestProperties(t *testing.T) {
	cmd := serve.NewCommand(context.Background(), zap.NewNop())

	assert.Equal(t, "serve", cmd.Use)
	assert.ElementsMatch(t, []string{"s", "server"}, cmd.Aliases)
	assert.NotNil(t, cmd.RunE)
}

func TestFlags(t *testing.T) {
	cmd := serve.NewCommand(context.Background(), zap.NewNop())

	cases := []struct {
		giveName      string
		wantShorthand string
		wantDefault   string
	}{
		{giveName: "listen", wantShorthand: "l", wantDefault: "0.0.0.0"},
		{giveName: "port", wantShorthand: "p", wantDefault: "8080"},
		{giveName: "prefix", wantShorthand: "x", wantDefault: "proxy"},
		{giveName: "proxy-request-timeout", wantShorthand: "", wantDefault: "30s"},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.giveName, func(t *testing.T) {
			flag := cmd.Flag(tt.giveName)

			if flag == nil {
				assert.Failf(t, "flag not found", "flag [%s] was not found", tt.giveName)

				return
			}

			assert.Equal(t, tt.wantShorthand, flag.Shorthand)
			assert.Equal(t, tt.wantDefault, flag.DefValue)
		})
	}
}

func TestSuccessfulFlagsPreparing(t *testing.T) {
	cmd := serve.NewCommand(context.Background(), zap.NewNop())
	// cmd.SetArgs([]string{"--any-flag", "any-value"})

	var executed bool

	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true

		return nil
	}

	output := capturer.CaptureOutput(func() {
		assert.NoError(t, cmd.Execute())
	})

	assert.Empty(t, output)
	assert.True(t, executed)
}

func TestFlagsWorkingWithoutCommandExecution(t *testing.T) {
	for _, tt := range []struct {
		name             string
		giveEnv          map[string]string
		giveArgs         []string
		wantErrorStrings []string
	}{
		{
			name: "Listen Flag Wrong Argument",
			giveArgs: []string{
				"-l", "256.256.256.256", // 255 is max
			},
			wantErrorStrings: []string{"wrong IP address", "256.256.256.256"},
		},
		{
			name:    "Listen Flag Wrong Env Value",
			giveEnv: map[string]string{"LISTEN_ADDR": "256.256.256.256"}, // 255 is max
			giveArgs: []string{
				"-l", "0.0.0.0", // `-l` flag must be ignored
			},
			wantErrorStrings: []string{"wrong IP address", "256.256.256.256"},
		},
		{
			name: "Port Flag Wrong Argument",
			giveArgs: []string{
				"-p", "65536", // 65535 is max
			},
			wantErrorStrings: []string{"invalid argument", "65536", "value out of range"},
		},
		{
			name:    "Port Flag Wrong Env Value",
			giveEnv: map[string]string{"LISTEN_PORT": "65536"}, // 65535 is max
			giveArgs: []string{
				"-p", "8090", // `-p` flag must be ignored
			},
			wantErrorStrings: []string{"wrong TCP port", "environment variable", "65536"},
		},
		{
			name: "Proxy Prefix Flag Wrong Argument",
			giveArgs: []string{
				"--prefix", "$$$", // invalid value
			},
			wantErrorStrings: []string{"wrong proxy prefix", "$$$"},
		},
		{
			name:    "Proxy Prefix Flag Wrong Env Value",
			giveEnv: map[string]string{"PROXY_PREFIX": "$$$"}, // invalid value
			giveArgs: []string{
				"--prefix", "foo", // valid value, but must be ignored
			},
			wantErrorStrings: []string{"wrong proxy prefix", "$$$"},
		},
		{
			name: "Proxy Request Timeout Flag Wrong Argument",
			giveArgs: []string{
				"--proxy-request-timeout", "1d", // invalid value
			},
			wantErrorStrings: []string{"invalid argument", "1d"},
		},
		{
			name:    "Proxy Request Timeout Flag Wrong Env Value",
			giveEnv: map[string]string{"PROXY_REQUEST_TIMEOUT": "1d"}, // invalid value
			giveArgs: []string{
				"--proxy-request-timeout", "1h", // valid value, but must be ignored
			},
			wantErrorStrings: []string{"wrong proxy request timeout", "1d"},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cmd := serve.NewCommand(context.Background(), zap.NewNop())
			cmd.SetArgs(tt.giveArgs)

			var executed bool

			cmd.RunE = func(*cobra.Command, []string) error {
				executed = true

				return nil
			}

			for k, v := range tt.giveEnv {
				assert.NoError(t, os.Setenv(k, v))
			}

			output := capturer.CaptureStderr(func() {
				assert.Error(t, cmd.Execute())
			})

			for k := range tt.giveEnv {
				assert.NoError(t, os.Unsetenv(k))
			}

			assert.False(t, executed)

			for _, want := range tt.wantErrorStrings {
				assert.Contains(t, output, want)
			}
		})
	}
}

func getRandomTCPPort(t *testing.T) (int, error) {
	t.Helper()

	// zero port means randomly (os) chosen port
	l, err := net.Listen("tcp", ":0") //nolint:gosec
	if err != nil {
		return 0, err
	}

	port := l.Addr().(*net.TCPAddr).Port

	if closingErr := l.Close(); closingErr != nil {
		return 0, closingErr
	}

	return port, nil
}

func checkTCPPortIsBusy(t *testing.T, port int) bool {
	t.Helper()

	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return true
	}

	_ = l.Close()

	return false
}

func startAndStopServer(t *testing.T, port int, args []string) string {
	t.Helper()

	var (
		output     string
		executedCh = make(chan struct{})
	)

	// start HTTP server
	go func(ch chan<- struct{}) {
		defer close(ch)

		output = capturer.CaptureStderr(func() {
			// create command with valid flags to run
			log, _ := zap.NewDevelopment()
			cmd := serve.NewCommand(context.Background(), log)
			cmd.SilenceUsage = true
			cmd.SetArgs(args)

			assert.NoError(t, cmd.Execute())
		})

		ch <- struct{}{}
	}(executedCh)

	portBusyCh := make(chan struct{})

	// check port "busy" (by HTTP server) state
	go func(ch chan<- struct{}) {
		defer close(ch)

		for i := 0; i < 2000; i++ {
			if checkTCPPortIsBusy(t, port) {
				ch <- struct{}{}

				return
			}

			<-time.After(time.Millisecond * 2)
		}

		t.Error("port opening timeout exceeded")
	}(portBusyCh)

	<-portBusyCh // wait for server starting

	// send OS signal for server stopping
	proc, err := os.FindProcess(os.Getpid())
	assert.NoError(t, err)
	assert.NoError(t, proc.Signal(syscall.SIGINT)) // send the signal

	<-executedCh // wait until server has been stopped

	return output
}

func TestSuccessfulCommandRunning(t *testing.T) {
	// get TCP port number for a test
	port, err := getRandomTCPPort(t)
	assert.NoError(t, err)

	output := startAndStopServer(t, port, []string{
		"--listen", "127.0.0.1",
		"--port", strconv.Itoa(port),
		"--prefix", "foo",
		"--proxy-request-timeout", "30s",
	})

	assert.Contains(t, output, "Server starting")
	assert.Contains(t, output, "Stopping by OS signal")
	assert.Contains(t, output, "Server stopping")
}
