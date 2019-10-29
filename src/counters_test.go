package main

import (
	"reflect"
	"testing"
)

func TestCounters_Increment(t *testing.T) {
	t.Parallel()

	var counters = NewCounters(map[string]int64{"foo": 1})

	if counters.Increment("foo") != 2 {
		t.Error("Existing value must be incremented by one")
	}

	if counters.Increment("bar") != 1 {
		t.Error("Non-existing counter value must be auto-initialized with zero and be incremented by one")
	}

	if counters.Increment("bar") != 2 {
		t.Error("Just created value must be incremented")
	}
}

func TestCounters_Decrement(t *testing.T) {
	t.Parallel()

	var counters = NewCounters(map[string]int64{"foo": 5})

	if counters.Decrement("foo") != 4 {
		t.Error("Value must be decremented by one")
	}

	if counters.Decrement("foo") != 3 {
		t.Error("Value must be decremented by one (again)")
	}

	if counters.Decrement("bar") != -1 {
		t.Error("Non-existing counter value must be auto-initialized with zero and be decremented by one")
	}
}

func TestCounters_Exists(t *testing.T) {
	t.Parallel()

	var counters = NewCounters(map[string]int64{"foo": 0})

	if !counters.Exists("foo") {
		t.Error("Counter named [foo] must be exists")
	}

	if counters.Exists("bar") {
		t.Error("Counter named [bar] must NOT exists")
	}
}

func TestCounters_Get(t *testing.T) {
	t.Parallel()

	var counters = NewCounters(map[string]int64{"foo": 1})

	if counters.Get("foo") != 1 {
		t.Error("Existing value must returns")
	}

	if counters.Get("bar") != 0 {
		t.Error("Non-existing counter value must returns zero")
	}
}

func TestCounters_Set(t *testing.T) {
	t.Parallel()

	var counters = NewCounters(nil)

	counters.Set("foo", 999)
	counters.Set("bar", -999)

	if counters.Get("foo") != 999 {
		t.Error("Wrong counter value set")
	}

	if counters.Get("bar") != -999 {
		t.Error("Wrong counter value set")
	}
}

func TestCounters_All(t *testing.T) {
	t.Parallel()

	var counters = NewCounters(map[string]int64{"foo": 1, "bar": 2})
	counters.Set("baz", 3)

	if !reflect.DeepEqual(counters.All(), map[string]int64{"foo": 1, "bar": 2, "baz": 3}) {
		t.Error("All must returns all values")
	}
}
