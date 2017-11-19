package ovirt_api

import "log"

// Logger logs messages
type Logger interface {
	// Info logs message with info severity
	Info(args ...string)
	// Debug logs message with debug severity
	Debug(args ...string)
	// Error logs message with error severity
	Error(args ...string)
}

type DefaultLogger struct {
}

func (l *DefaultLogger) Info(args ...string) {
	log.Println(args)
}

func (l *DefaultLogger) Debug(args ...string) {
	log.Println(args)
}

func (l *DefaultLogger) Error(args ...string) {
	log.Println(args)
}