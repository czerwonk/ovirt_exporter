package main

import "github.com/prometheus/common/log"

// PromLogger implements github.com/czerwonk/ovirt_api/Logger
type PromLogger struct {
}

// Info logs info messages
func (l *PromLogger) Info(args ...string) {
	log.Info(args)
}

// Debug logs debug messages
func (l *PromLogger) Debug(args ...string) {
	log.Debug(args)
}

// Error logs errors
func (l *PromLogger) Error(args ...string) {
	log.Error(args)
}
