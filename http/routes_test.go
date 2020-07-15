package http

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer_RegisterHandlers(t *testing.T) {
	t.Parallel()

	var s = NewServer(&ServerSettings{ProxyRoutePrefix: "foobar"})

	var cases = []struct {
		giveName         string
		wantPathTemplate string
		wantMethods      []string
	}{
		{
			giveName:         "index",
			wantPathTemplate: "/",
			wantMethods:      []string{http.MethodGet},
		},
		{
			giveName:         "ping",
			wantPathTemplate: "/ping",
			wantMethods:      []string{http.MethodGet},
		},
		{
			giveName:         "metrics",
			wantPathTemplate: "/metrics",
			wantMethods:      []string{http.MethodGet},
		},
		{
			giveName:         "proxy",
			wantPathTemplate: "/foobar/{uri:.*}",
			wantMethods: []string{
				http.MethodGet,
				http.MethodHead,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodOptions,
			},
		},
	}

	for _, tt := range cases {
		assert.Nil(t,
			s.Router.Get(tt.giveName),
			fmt.Sprintf("Handler for route [%s] must be not registered", tt.giveName),
		)
	}

	s.RegisterHandlers()

	for _, tt := range cases {
		t.Run(tt.giveName, func(t *testing.T) {
			route := s.Router.Get(tt.giveName)

			pathTemplate, pathTemplateErr := route.GetPathTemplate()
			assert.Nil(t, pathTemplateErr)
			assert.Equal(t, tt.wantPathTemplate, pathTemplate)

			routeMethods, routeMethodsErr := route.GetMethods()
			assert.Nil(t, routeMethodsErr)
			assert.Equal(t, tt.wantMethods, routeMethods)
		})
	}
}
