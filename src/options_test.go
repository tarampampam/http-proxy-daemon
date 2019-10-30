package main

import (
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

//func TestNewOptions(t *testing.T) {
//	t.Parallel()
//}
