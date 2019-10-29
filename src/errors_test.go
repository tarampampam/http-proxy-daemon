package main

import (
	"encoding/json"
	"testing"
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

		if !got.IsError {
			t.Errorf("Flag [%s] must be true", "IsError")
		}

		if got.Code != tt.code {
			t.Error("Wrong error code set: ", got.Code)
		}

		if got.Message != tt.message {
			t.Error("Wrong error message set: ", got.Message)
		}

		if got.Error() != got.Message {
			t.Errorf("Wrong method [%s] implementation", "Error()")
		}
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

	exp := `{"code":123,"error":true,"message":"foo"}`

	j, _ := json.Marshal(err)

	if string(j) != exp {
		t.Errorf("Wrong JSON encoding. Expected: [%s], got: [%s]", exp, string(j))
	}
}
