package ovirt_api

// Logger logs messages
type Logger interface {
	// Info logs message with info severity
	Info(args ...string)
	// Debug logs message with debug severity
	Debug(args ...string)
	// Error logs message with error severity
	Error(args ...string)
}
