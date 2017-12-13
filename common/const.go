package common

const DefaultCacheDir = "findercache"

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

const (
	ConfigEventPrefix          = "config_"
	ServiceEventPrefix         = "service_"
	ServiceConfEventPrefix     = "service_conf_"
	ServiceProviderEventPrefix = "service_provider_"
	ServiceConsumerEventPrefix = "service_consumer_"
)
