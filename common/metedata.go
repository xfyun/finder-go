package common

import "time"

type ServiceMeteData struct {
	Project string
	Group   string
	Service string
	Version string
}

type BootConfig struct {
	CompanionUrl     string
	CachePath        string
	TickerDuration   time.Duration
	ZkSessionTimeout time.Duration
	ZkConnectTimeout time.Duration
	ZkMaxSleepTime   time.Duration
	ZkMaxRetryNum    int
	MeteData         ServiceMeteData
}

type ZkInfo struct {
	ZkAddr          []string
	ConfigRootPath  string
	ServiceRootPath string
}

type Config struct {
	Name string
	File []byte
}

type ServiceItem struct {
	Addr    string
	Weight  int
	IsValid bool
}

type Service struct {
	Name       string
	ServerList []ServiceItem
	Extra      map[string]string
}
