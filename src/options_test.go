package main

import (
	"bytes"
	"github.com/jessevdk/go-flags"
	"io"
	"reflect"
	"testing"
)

func TestNewOptions(t *testing.T) {
	type args struct {
		onExit OptionsExitFunc
	}
	tests := []struct {
		name       string
		args       args
		wantStdOut string
		wantStdErr string
		want       *Options
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdOut := &bytes.Buffer{}
			stdErr := &bytes.Buffer{}
			got := NewOptions(stdOut, stdErr, tt.args.onExit)
			if gotStdOut := stdOut.String(); gotStdOut != tt.wantStdOut {
				t.Errorf("NewOptions() gotStdOut = %v, want %v", gotStdOut, tt.wantStdOut)
			}
			if gotStdErr := stdErr.String(); gotStdErr != tt.wantStdErr {
				t.Errorf("NewOptions() gotStdErr = %v, want %v", gotStdErr, tt.wantStdErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptions_Check(t *testing.T) {
	type fields struct {
		Address     string
		Port        int
		ProxyPrefix string
		ShowVersion bool
		stdOut      io.Writer
		stdErr      io.Writer
		onExit      OptionsExitFunc
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Options{
				Address:     tt.fields.Address,
				Port:        tt.fields.Port,
				ProxyPrefix: tt.fields.ProxyPrefix,
				ShowVersion: tt.fields.ShowVersion,
				stdOut:      tt.fields.stdOut,
				stdErr:      tt.fields.stdErr,
				onExit:      tt.fields.onExit,
			}
			got, err := o.Check()
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Check() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptions_Parse(t *testing.T) {
	type fields struct {
		Address     string
		Port        int
		ProxyPrefix string
		ShowVersion bool
		stdOut      io.Writer
		stdErr      io.Writer
		onExit      OptionsExitFunc
	}
	tests := []struct {
		name   string
		fields fields
		want   *flags.Parser
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Options{
				Address:     tt.fields.Address,
				Port:        tt.fields.Port,
				ProxyPrefix: tt.fields.ProxyPrefix,
				ShowVersion: tt.fields.ShowVersion,
				stdOut:      tt.fields.stdOut,
				stdErr:      tt.fields.stdErr,
				onExit:      tt.fields.onExit,
			}
			if got := o.Parse(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
