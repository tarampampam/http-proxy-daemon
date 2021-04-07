package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, "LISTEN_ADDR", string(ListenAddr))
	assert.Equal(t, "LISTEN_PORT", string(ListenPort))
	assert.Equal(t, "PROXY_PREFIX", string(ProxyRoutePrefix))
	assert.Equal(t, "PROXY_REQUEST_TIMEOUT", string(ProxyRequestTimeout))
}

func TestEnvVariable_Lookup(t *testing.T) {
	cases := []struct {
		giveEnv envVariable
	}{
		{giveEnv: ListenAddr},
		{giveEnv: ListenPort},
		{giveEnv: ProxyRoutePrefix},
		{giveEnv: ProxyRequestTimeout},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.giveEnv.String(), func(t *testing.T) {
			assert.NoError(t, os.Unsetenv(tt.giveEnv.String())) // make sure that env is unset for test

			defer func() { assert.NoError(t, os.Unsetenv(tt.giveEnv.String())) }()

			value, exists := tt.giveEnv.Lookup()
			assert.False(t, exists)
			assert.Empty(t, value)

			assert.NoError(t, os.Setenv(tt.giveEnv.String(), "foo"))

			value, exists = tt.giveEnv.Lookup()
			assert.True(t, exists)
			assert.Equal(t, "foo", value)
		})
	}
}
