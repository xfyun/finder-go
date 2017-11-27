package finder

import "finder-go/zk"

type ServiceMeta struct {
	project string
	group   string
	service string
	version string
}

type BootConfig struct {
	companionUrl string
	serviceMeta  ServiceMeta
}

type FinderManager struct {
	configFinder  *ConfigFinder
	serviceFinder *ServiceFinder
	zkManager     *zk.ZkManager
}

func NewFinder(config BootConfig) (*FinderManager, error) {
	fm := new(FinderManager)
	fm.configFinder = new(ConfigFinder)
	fm.serviceFinder = new(ServiceFinder)

	return fm, nil
}

func main() {
	config := BootConfig{
		companionUrl:"http://xxx.xxx.xx/",
		serviceMeta:ServiceMeta{
			project:"7s",
			group:"set1",
			service:"sis",
			version:"1.0.1",
		},
	}

	f, err := NewFinder(config)

	if (err != nil) {

	}

	f.configFinder.UseConfig("default.cfg", true, onCfgUpdateEvent)
	f.serviceFinder.RegisterService("sis", "172.27.0.16:9090")
}

func onCfgUpdateEvent(c Config) int {
	return ConfigSuccess
}
