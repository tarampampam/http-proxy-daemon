package counters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryCounter_Increment(t *testing.T) {
	t.Parallel()

	var counters = NewInMemoryCounters(map[string]int64{"foo": 1})

	counters.Increment("foo")
	assert.Equal(t,
		int64(2),
		counters.Get("foo"),
		"Existing value must be incremented by one",
	)

	counters.Increment("bar")
	assert.Equal(t,
		int64(1),
		counters.Get("bar"),
		"Non-existing counters value must be auto-initialized with zero and be incremented by one",
	)

	counters.Increment("bar")
	assert.Equal(t,
		int64(2),
		counters.Get("bar"),
		"Just created value must be incremented",
	)
}

func TestInMemoryCounter_Decrement(t *testing.T) {
	t.Parallel()

	var counters = NewInMemoryCounters(map[string]int64{"foo": 5})

	counters.Decrement("foo")
	assert.Equal(t,
		int64(4),
		counters.Get("foo"),
		"Existing value must be incremented by one",
	)

	counters.Decrement("foo")
	assert.Equal(t,
		int64(3),
		counters.Get("foo"),
		"Value must be decremented by one (again)",
	)

	counters.Decrement("bar")
	assert.Equal(t,
		int64(-1),
		counters.Get("bar"),
		"Non-existing counters value must be auto-initialized with zero and be decremented by one",
	)
}

func TestInMemoryCounter_Get(t *testing.T) {
	t.Parallel()

	var counters = NewInMemoryCounters(map[string]int64{"foo": 1})

	assert.Equal(t, int64(1), counters.Get("foo"), "Existing value must returns")
	assert.Equal(t, int64(0), counters.Get("bar"), "Non-existing counters value must returns zero")
}

func TestInMemoryCounter_Set(t *testing.T) {
	t.Parallel()

	var counters = NewInMemoryCounters(nil)

	counters.Set("foo", 999)
	counters.Set("bar", -999)

	assert.Equal(t, int64(999), counters.Get("foo"), "Wrong counters value is set")
	assert.Equal(t, int64(-999), counters.Get("bar"), "Wrong counters value is set")
}
