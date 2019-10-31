package main

import (
	"errors"
	"io/ioutil"
	"os"
	"reflect"
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
		options    *Options
		wantError  error
		wantResult bool
	}{
		{
			options: &Options{
				ProxyPrefix: "$%^", // <-- !!!
				Address:     "127.0.0.1",
				Port:        8080,
			},
			wantError:  errors.New("wrong prefix passed"),
			wantResult: false,
		},
		{
			options: &Options{
				ProxyPrefix: "  ", // <-- !!!
				Address:     "127.0.0.1",
				Port:        8080,
			},
			wantError:  errors.New("wrong prefix passed"),
			wantResult: false,
		},
		{
			options: &Options{
				ProxyPrefix: "proxy",
				Address:     "foo", // <-- !!!
				Port:        8080,
			},
			wantError:  errors.New("wrong address to listen on"),
			wantResult: false,
		},
		{
			options: &Options{
				ProxyPrefix: "proxy",
				Address:     "127.0.0.1",
				Port:        0, // <-- !!!
			},
			wantError:  errors.New("wrong port number"),
			wantResult: false,
		},
		{
			options: &Options{
				ProxyPrefix: "proxy",
				Address:     "127.0.0.1",
				Port:        65536, // <-- !!!
			},
			wantError:  errors.New("wrong port number"),
			wantResult: false,
		},
		{
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
			options: &Options{ // Success Case
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
	}
}
