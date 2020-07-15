package errors

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		message string
		code    int
	}{
		{
			message: "foo",
			code:    100,
		},
	}
	for _, tt := range tests {
		got := NewServerError(tt.code, tt.message)

		assert.True(t, got.IsError)
		assert.Equal(t, tt.code, got.Code)
		assert.Equal(t, tt.message, got.Message)
		assert.Equal(t, tt.message, got.Error())
	}
}

func TestServerError_JsonCasting(t *testing.T) {
	t.Parallel()

	err := ServerError{
		error:   nil,
		Code:    123,
		IsError: true,
		Message: "foo",
	}

	asJSON, _ := json.Marshal(err)

	assert.JSONEq(t, `{"code":123,"error":true,"message":"foo"}`, string(asJSON))
	assert.JSONEq(t, `{"code":123,"error":true,"message":"foo"}`, string(err.ToJSON()))
}
