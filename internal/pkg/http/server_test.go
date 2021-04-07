package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/config"
	"go.uber.org/zap"
)

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

func TestServer_StartAndStop(t *testing.T) {
	port, err := getRandomTCPPort(t)
	assert.NoError(t, err)

	srv := NewServer(zap.NewNop())

	assert.False(t, checkTCPPortIsBusy(t, port))

	go func() {
		startingErr := srv.Start("", uint16(port))

		if !errors.Is(startingErr, http.ErrServerClosed) {
			assert.NoError(t, startingErr)
		}
	}()

	for i := 0; ; i++ {
		if !checkTCPPortIsBusy(t, port) {
			if i > 100 {
				t.Error("too many attempts for server start checking")
			}

			<-time.After(time.Microsecond * 10)
		} else {
			break
		}
	}

	assert.True(t, checkTCPPortIsBusy(t, port))
	assert.NoError(t, srv.Stop(context.Background()))
	assert.False(t, checkTCPPortIsBusy(t, port))
}

func TestServer_Register(t *testing.T) {
	var routes = []struct {
		name    string
		route   string
		methods []string
	}{
		{name: "proxy", route: "/foo/{uri:.*}"},
		{name: "index", route: "/", methods: []string{http.MethodGet}},
		{name: "metrics", route: "/metrics", methods: []string{http.MethodGet}},
		{name: "ready", route: "/ready", methods: []string{http.MethodGet, http.MethodHead}},
		{name: "live", route: "/live", methods: []string{http.MethodGet, http.MethodHead}},
	}

	srv := NewServer(zap.NewNop())

	router := srv.router // dirty hack, yes, i know

	for _, r := range routes {
		assert.Nil(t, router.Get(r.name))
	}

	cfg := config.Config{}
	cfg.Proxy.Prefix = "foo"

	// call register fn
	assert.NoError(t, srv.Register(context.Background(), cfg))

	for _, r := range routes {
		route, _ := router.Get(r.name).GetPathTemplate()
		assert.Equal(t, r.route, route)

		if len(r.methods) > 0 {
			methods, _ := router.Get(r.name).GetMethods()
			assert.Equal(t, r.methods, methods)
		}
	}
}

func TestServer_RegisterWithoutProxyPrefix(t *testing.T) {
	srv := NewServer(zap.NewNop())

	assert.EqualError(t, srv.Register(context.Background(), config.Config{}), "empty proxy prefix")
}
