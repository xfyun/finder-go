package common

type ServiceChangedHandler interface {
	OnServiceInstanceConfigChanged(name string, addr string, config *ServiceInstanceConfig) bool
	OnServiceConfigChanged(name string, config *ServiceConfig) bool
	OnServiceInstanceChanged(name string, instances []*ServiceInstance) bool
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
