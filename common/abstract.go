package common

import "log"

type ServiceChangedHandler interface {
	OnServiceInstanceConfigChanged(name string, addr string, config *ServiceInstanceConfig) bool
	OnServiceConfigChanged(name string, config *ServiceConfig) bool
	OnServiceInstanceChanged(name string, eventList []*ServiceInstanceChangedEvent) bool
}

type ConfigChangedHandler interface {
	OnConfigFileChanged(config *Config) bool
}

type InternalServiceChangedHandler interface {
	OnServiceInstanceConfigChanged(name string, addr string, data []byte)
	OnServiceConfigChanged(name string, data []byte)
	OnServiceInstanceChanged(name string, addrList []string)
}

type InternalConfigChangedHandler interface {
	OnConfigFileChanged(name string, data []byte)
}

type Logger interface {
	Info(v ...interface{})
	Debug(v ...interface{})
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
