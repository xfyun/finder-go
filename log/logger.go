package log

import "log"

type Logger interface {
	Info(v ...interface{})
	Debug(v ...interface{})
	Error(v ...interface{})
	Infof(fmt string, v ...interface{})
	Debugf(fmt string, v ...interface{})
	Errorf(fmt string, v ...interface{})
}

type DefaultLogger struct {
}

func NewDefaultLogger() Logger {
	return &DefaultLogger{}
}

func (l *DefaultLogger) Info(v ...interface{}) {
	log.Println(v)
}

func (l *DefaultLogger) Debug(v ...interface{}) {
	log.Println(v)
}

func (l *DefaultLogger) Error(v ...interface{}) {
	log.Println(v)
}

func (l *DefaultLogger) Infof(fmt string, v ...interface{}) {
	log.Printf(fmt, v)
}

func (l *DefaultLogger) Debugf(fmt string, v ...interface{}) {
	log.Printf(fmt, v)
}

func (l *DefaultLogger) Errorf(fmt string, v ...interface{}) {
	log.Printf(fmt, v)
}
