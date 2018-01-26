package api

import "log"

// Logger logs messages
type Logger interface {
	// Info logs message with info severity
	Infof(format string, args ...string)
	// Debug logs message with debug severity
	Debugf(format string, args ...string)
	// Error logs message with error severity
	Errorf(format string, args ...string)
}

// DefaultLogger is the default impplemetation of Logger interface using golangs log package
type defaultLogger struct {
}

// Info implements Logger interfae
func (l *defaultLogger) Infof(format string, args ...string) {
	log.Printf(format, args)
}

// Debug implements Logger interfae
func (l *defaultLogger) Debugf(format string, args ...string) {
	log.Printf(format, args)
}

// Error implements Logger interfae
func (l *defaultLogger) Errorf(format string, args ...string) {
	log.Printf(format, args)
}
