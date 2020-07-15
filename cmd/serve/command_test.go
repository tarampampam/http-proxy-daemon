package serve

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_address_IsValidValue(t *testing.T) {
	tests := []struct {
		giveAddr  string
		wantError bool
	}{
		{giveAddr: "1.1.1.1", wantError: false},
		{giveAddr: "foobar", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.giveAddr, func(t *testing.T) {
			err := new(address).IsValidValue(tt.giveAddr)

			if tt.wantError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_address_String(t *testing.T) {
	assert.Equal(t, "0.0.0.0", address("0.0.0.0").String())
}

func Test_port_String(t *testing.T) {
	assert.Equal(t, "123", port(123).String())
}

func Test_prefix_IsValidValue(t *testing.T) {
	tests := []struct {
		givePrefix string
		wantError  bool
	}{
		{givePrefix: "foo", wantError: false},
		{givePrefix: "foobar", wantError: false},
		{givePrefix: "foo/bar", wantError: false},
		{givePrefix: "foo_bar", wantError: false},
		{givePrefix: "foo-bar-", wantError: false},
		{givePrefix: "foo123", wantError: false},
		{givePrefix: "foo_123", wantError: false},
		{givePrefix: "foo bar", wantError: true},
		{givePrefix: " foo", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.givePrefix, func(t *testing.T) {
			err := new(prefix).IsValidValue(tt.givePrefix)

			if tt.wantError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_prefix_String(t *testing.T) {
	assert.Equal(t, "foobar", prefix("foobar").String())
}

func TestCommand_Execute(t *testing.T) {
	t.Skip("Not implemented yet")
}
