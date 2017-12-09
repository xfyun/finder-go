package common

const DefaultCacheDir = "findercache"

type ServiceChangedHandler interface {
	OnServiceInstanceConfigChanged(name string, instance string, config ServiceItemConfig) bool
	OnServiceConfigChanged(name string, config ServiceConfig) bool
	OnServiceInstanceChanged(name string, instances []ServiceItem) bool
}

type ConfigChangedHandler interface {
	OnConfigFileChanged(config Config) bool
}

type InternalServiceChangedHandler interface {
	OnServiceInstanceConfigChanged(name string, instance string, data []byte)
	OnServiceConfigChanged(name string, data []byte)
	OnServiceInstanceChanged(name string, instances []string)
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
