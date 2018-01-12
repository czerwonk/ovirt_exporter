package api

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

// DefaultLogger is the default impplemetation of Logger interface using golangs log package
type DefaultLogger struct {
}

// Info implements Logger interfae
func (l *DefaultLogger) Info(args ...string) {
	log.Println(args)
}

// Debug implements Logger interfae
func (l *DefaultLogger) Debug(args ...string) {
	log.Println(args)
}

// Error implements Logger interfae
func (l *DefaultLogger) Error(args ...string) {
	log.Println(args)
}
