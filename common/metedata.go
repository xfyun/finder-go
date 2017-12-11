package common

import "time"

type ServiceMeteData struct {
	Project string
	Group   string
	Service string
	Version string
	Address string
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

type ServiceInstanceConfig struct {
	Weight  int  `json:"weight"`
	IsValid bool `json:"is_valid"`
}

type ServiceInstance struct {
	Addr   string
	Config *ServiceInstanceConfig
}

type ServiceConfig struct {
	ProxyMode       string `json:"proxy_mode"`
	LoadBalanceMode string `json:"lb_mode"`
}

type Service struct {
	Name       string
	ServerList []ServiceInstance
	Config     *ServiceConfig
}
