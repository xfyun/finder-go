package finder

import "log"

type ServiceChangedHandler interface {

	//服务实例上的配置信息发生变化
	OnServiceInstanceConfigChanged(name string,apiVersion string, addr string, config *ServiceInstanceConfig) bool
	//服务整体配置信息发生变化
	OnServiceConfigChanged(name string,apiVersion string,  config *ServiceConfig) bool
	//服务实例发生变化
	OnServiceInstanceChanged(name string, apiVersion string, eventList []*ServiceInstanceChangedEvent) bool
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
	Error(v ...interface{})
	Infof(fmt string, v ...interface{})
	Debugf(fmt string, v ...interface{})
	Errorf(fmt string, v ...interface{})
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
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
