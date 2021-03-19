package main

import "github.com/prometheus/common/log"

// PromLogger implements github.com/czerwonk/ovirt_api/Logger
type PromLogger struct {
}

// Infof logs info messages
func (l *PromLogger) Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Debugf logs debug messages
func (l *PromLogger) Debugf(format string, args ...interface{}) {
	if *debug {
		log.Debugf(format, args...)
	}
}

// Errorf logs errors
func (l *PromLogger) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}
