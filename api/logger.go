package api

type Logger interface {
	Info(args ...string)
	Debug(args ...string)
	Error(args ...string)
}
