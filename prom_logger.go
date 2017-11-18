package main

import "github.com/prometheus/common/log"

type PromLogger struct {
}

func (l *PromLogger) Info(args ...string) {
	log.Info(args)
}

func (l *PromLogger) Debug(args ...string) {
	log.Debug(args)
}

func (l *PromLogger) Error(args ...string) {
	log.Error(args)
}
