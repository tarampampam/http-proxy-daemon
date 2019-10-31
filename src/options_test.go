package main

import (
	"errors"
	"github.com/jessevdk/go-flags"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestOptions_StructTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		element         func() reflect.StructField
		wantShort       string
		wantLong        string
		wantEnv         string
		wantDefault     string
		wantDescription string
	}{
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(Options{}).FieldByName("Address")
				return field
			},
			wantShort:       "l",
			wantLong:        "listen",
			wantEnv:         "LISTEN_ADDR",
			wantDefault:     "0.0.0.0",
			wantDescription: "Address (IP) to listen on",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(Options{}).FieldByName("Port")
				return field
			},
			wantShort:       "p",
			wantLong:        "port",
			wantEnv:         "LISTEN_PORT",
			wantDefault:     "8080",
			wantDescription: "TCP port number",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(Options{}).FieldByName("ProxyPrefix")
				return field
			},
			wantShort:       "x",
			wantLong:        "prefix",
			wantEnv:         "PROXY_PREFIX",
			wantDefault:     "proxy",
			wantDescription: "Proxy route prefix",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(Options{}).FieldByName("ShowVersion")
				return field
			},
			wantShort:       "V",
			wantLong:        "version",
			wantDescription: "Show version and exit",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(Options{}).FieldByName("TslCertFile")
				return field
			},
			wantLong:        "tsl-cert",
			wantEnv:         "TSL_CERT",
			wantDescription: "TSL certificate file path",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(Options{}).FieldByName("TslKeyFile")
				return field
			},
			wantLong:        "tsl-key",
			wantEnv:         "TSL_KEY",
			wantDescription: "TSL key file path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.wantDescription, func(t *testing.T) {
			el := tt.element()
			if tt.wantShort != "" {
				value, _ := el.Tag.Lookup("short")
				if value != tt.wantShort {
					t.Errorf("Wrong value for 'short' tag. Want: %v, got: %v", tt.wantShort, value)
				}
			}

			if tt.wantLong != "" {
				value, _ := el.Tag.Lookup("long")
				if value != tt.wantLong {
					t.Errorf("Wrong value for 'long' tag. Want: %v, got: %v", tt.wantLong, value)
				}
			}

			if tt.wantEnv != "" {
				value, _ := el.Tag.Lookup("env")
				if value != tt.wantEnv {
					t.Errorf("Wrong value for 'env' tag. Want: %v, got: %v", tt.wantEnv, value)
				}
			}

			if tt.wantDefault != "" {
				value, _ := el.Tag.Lookup("default")
				if value != tt.wantDefault {
					t.Errorf("Wrong value for 'default' tag. Want: %v, got: %v", tt.wantDefault, value)
				}
			}

			if tt.wantDescription != "" {
				value, _ := el.Tag.Lookup("description")
				if value != tt.wantDescription {
					t.Errorf("Wrong value for 'description' tag. Want: %v, got: %v", tt.wantDescription, value)
				}
			}
		})
	}
}

func TestOptions_Parse(t *testing.T) {
	t.Parallel()

	// Make args backup
	origArgs := make([]string, 0)
	origArgs = append(origArgs, os.Args...)

	// Restore args
	defer func() {
		os.Args = make([]string, 0)
		os.Args = append(os.Args, origArgs...)
	}()

	var (
		exited   bool
		exitCode int
		exitFunc OptionsExitFunc = func(code int) {
			exited = true
			exitCode = code
		}
		errLog  = log.New(&FakeWriter{}, "", 0)
		stdLog  = log.New(&FakeWriter{}, "", 0)
		options = &Options{
			ProxyPrefix: "proxy",
			Address:     "127.0.0.1",
			Port:        8080,
			onExit:      exitFunc,
			errLog:      errLog,
			stdLog:      stdLog,
			parseFlags:  flags.PassDoubleDash | flags.HelpFlag,
		}
	)

	tests := []struct {
		name            string
		options         *Options
		osArgs          []string
		wantExit        bool
		wantExitCode    int
		wantStdMessages []string
		wantErrMessages []string
	}{
		{
			name:         "Unsupported argument",
			options:      options,
			osArgs:       []string{"app", "-@"},
			wantExit:     true,
			wantExitCode: 1,
			wantStdMessages: []string{
				"Usage:", "Application Options:", "Help Options:",
			},
			// @todo: How test this? flags uses direct writing in os.Stderr
			//wantErrMessages: []string{"unknown flag", "@"},
		},
		{
			name:         "Show help",
			options:      options,
			osArgs:       []string{"app", "-h"},
			wantExit:     true,
			wantExitCode: 0,
			// @todo: How test this? flags uses direct writing in os.Stdout
			//wantStdMessages: []string{"Usage:", "Application Options:", "Help Options:"},
		},
		{
			name:            "Version requested",
			options:         options,
			osArgs:          []string{"app", "-V"},
			wantExit:        true,
			wantExitCode:    0,
			wantStdMessages: []string{"Version", VERSION},
		},
		{
			name:            "Known argument with wrong value",
			options:         options,
			osArgs:          []string{"app", "-p", "999999999999"},
			wantExit:        true,
			wantExitCode:    1,
			wantErrMessages: []string{"wrong port number"},
		},
		{
			name:         "All ok",
			options:      options,
			osArgs:       []string{"app"},
			wantExit:     false,
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = make([]string, 0)
			os.Args = append(os.Args, tt.osArgs...)

			tt.options.Parse()

			if tt.wantExit {
				if !exited {
					t.Error("exit callback was not called")
				}
				if exitCode != tt.wantExitCode {
					t.Errorf("wrong exit code: want %v, got %v", tt.wantExitCode, exitCode)
				}
			}

			errMessages := options.errLog.Writer().(*FakeWriter).ToStringAndClean()
			for _, wantErrEntry := range tt.wantErrMessages {
				if !strings.Contains(errMessages, wantErrEntry) {
					t.Errorf("Expected error message entry [%s] not found in: %v", wantErrEntry, errMessages)
				}
			}

			stdMessages := options.stdLog.Writer().(*FakeWriter).ToStringAndClean()
			for _, wantStdEntry := range tt.wantStdMessages {
				if !strings.Contains(stdMessages, wantStdEntry) {
					t.Errorf("Expected regular message entry [%s] not found in: %v", wantStdEntry, stdMessages)
				}
			}

			// reset state
			exited = false
			exitCode = -1
			options.errLog.Writer().(*FakeWriter).CleanBuf()
			options.stdLog.Writer().(*FakeWriter).CleanBuf()
		})
	}
}

func TestOptions_Check(t *testing.T) {
	t.Parallel()

	file, _ := ioutil.TempFile("", "unit-test-")
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
		if err := os.Remove(file.Name()); err != nil {
			panic(err)
		}
	}()

	tests := []struct {
		name       string
		options    *Options
		wantError  error
		wantResult bool
	}{
		{
			name: "Wrong proxy prefix",
			options: &Options{
				ProxyPrefix: "$%^", // <-- !!!
				Address:     "127.0.0.1",
				Port:        8080,
			},
			wantError:  errors.New("wrong prefix passed"),
			wantResult: false,
		},
		{
			name: "Empty (spaces only) in proxy prefix",
			options: &Options{
				ProxyPrefix: "  ", // <-- !!!
				Address:     "127.0.0.1",
				Port:        8080,
			},
			wantError:  errors.New("wrong prefix passed"),
			wantResult: false,
		},
		{
			name: "Wrong address",
			options: &Options{
				ProxyPrefix: "proxy",
				Address:     "foo", // <-- !!!
				Port:        8080,
			},
			wantError:  errors.New("wrong address to listen on"),
			wantResult: false,
		},
		{
			name: "Wrong port (less then min)",
			options: &Options{
				ProxyPrefix: "proxy",
				Address:     "127.0.0.1",
				Port:        0, // <-- !!!
			},
			wantError:  errors.New("wrong port number"),
			wantResult: false,
		},
		{
			name: "Wrong port (over max)",
			options: &Options{
				ProxyPrefix: "proxy",
				Address:     "127.0.0.1",
				Port:        65536, // <-- !!!
			},
			wantError:  errors.New("wrong port number"),
			wantResult: false,
		},
		{
			name: "Wrong TSL cert path",
			options: &Options{
				ProxyPrefix: "proxy",
				Address:     "127.0.0.1",
				Port:        8080,
				TslCertFile: "/foo/bar", // <-- !!!
				TslKeyFile:  file.Name(),
			},
			wantError:  errors.New("wrong TSL certificate file path"),
			wantResult: false,
		},
		{
			name: "Wrong TSL key file path",
			options: &Options{
				ProxyPrefix: "proxy",
				Address:     "127.0.0.1",
				Port:        8080,
				TslCertFile: file.Name(),
				TslKeyFile:  "/foo/bar", // <-- !!!
			},
			wantError:  errors.New("wrong TSL key file path"),
			wantResult: false,
		},
		{
			name: "Success case",
			options: &Options{
				ProxyPrefix: "proxy",
				Address:     "127.0.0.1",
				Port:        8080,
				TslCertFile: file.Name(),
				TslKeyFile:  file.Name(),
			},
			wantError:  nil,
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.options.Check()

			if tt.wantError != nil {
				if err == nil {
					t.Fatalf("expected error not returned. want: %v", tt.wantError)
				}
				if err.(error).Error() != tt.wantError.Error() {
					t.Errorf("unexpected error returned: want %v, got %v", tt.wantError, err)
				}
			}

			if res != tt.wantResult {
				t.Errorf("wrong result returned: want %v, got %v", tt.wantResult, res)
			}
		})
	}
}

func TestNewOptions(t *testing.T) {
	t.Parallel()

	compare := func(h1, h2 interface{}) bool {
		t.Helper()
		return reflect.ValueOf(h1).Pointer() == reflect.ValueOf(h2).Pointer()
	}

	var (
		onExit OptionsExitFunc = func(code int) {}
		errLog                 = log.New(&FakeWriter{}, "", 0)
		stdLog                 = log.New(&FakeWriter{}, "", 0)
		o                      = NewOptions(stdLog, errLog, onExit)
	)

	if !compare(o.onExit, onExit) {
		t.Error("Wrong onExit handle set")
	}
	if !compare(o.errLog, errLog) {
		t.Error("Wrong errLog set")
	}
	if !compare(o.stdLog, stdLog) {
		t.Error("Wrong stdLog set")
	}
}
