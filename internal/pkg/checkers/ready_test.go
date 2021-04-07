package checkers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/checkers"
)

func TestReadyChecker_Check(t *testing.T) {
	assert.NoError(t, checkers.NewReadyChecker().Check())
}
