package main

import (
	"io"
	"io/ioutil"
	"log"
	"testing"
)

func NewDeferredLoggerForTest(out io.Writer) (*DeferredLogger, *log.Logger) {
	if out == nil {
		out = ioutil.Discard
	}
	l := log.New(out, "", 0)
	df := NewDeferredLogger(l)

	return df, l
}

func TestNewDeferredLogger(t *testing.T) {
	t.Parallel()

	df, l := NewDeferredLoggerForTest(nil)

	if len(df.stack) != 0 {
		t.Errorf("Got unexprected stack: %v", df.stack)
	}

	if df.log != l {
		t.Errorf("Got unexpected logger: %v", df.log)
	}
}

func TestDeferredLogger_Add(t *testing.T) {
	t.Parallel()

	df, _ := NewDeferredLoggerForTest(nil)

	df.Add("one")

	if len(df.stack) != 1 {
		t.Errorf("Wrong stack size. Want: 1")
	}

	df.Add("two")

	if len(df.stack) != 2 {
		t.Errorf("Wrong stack size. Want: 2")
	}
}

func TestDeferredLogger_Flush(t *testing.T) {
	t.Parallel()

	w := &FakeWriter{}
	df, _ := NewDeferredLoggerForTest(w)

	df.Flush("")

	if data := w.ToStringAndClean(); data != "" {
		t.Errorf("On empty stack non-empty string returned: %v", data)
	}

	df.Add("one")
	df.Add("two")

	df.Flush("|")

	if data := w.ToStringAndClean(); data != "one|two\n" {
		t.Errorf("On non-empty stack empty string returned: %v", data)
	}
}
