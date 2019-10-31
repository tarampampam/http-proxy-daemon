package main

import (
	"log"
	"strings"
)

type DeferredLogger struct {
	stack []string
	log   *log.Logger
}

func NewDeferredLogger(log *log.Logger) *DeferredLogger {
	return &DeferredLogger{
		log: log,
	}
}

func (l *DeferredLogger) Add(message string) {
	l.stack = append(l.stack, message)
}

func (l *DeferredLogger) Flush(delimiter string) {
	if len(l.stack) > 0 {
		l.log.Println(strings.Join(l.stack, delimiter))
		l.stack = make([]string, 0)
	}
}
