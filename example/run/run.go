package main

import (
	"flag"
	"fmt"
	"github.com/xfyun/finder-go"
	common "github.com/xfyun/finder-go/common"
	"time"
)

func main() {
	flag.Parse()
	fd, err := finder.NewFinderWithLogger(common.BootConfig{
		CompanionUrl:  "http://10.1.87.69:6868",
		CachePath:     "./findercache",
		CacheConfig:   false,
		CacheService:  false,
		ExpireTimeout: 5 * time.Second,
		MeteData: &common.ServiceMeteData{
			Project: "AIPaaS",
			Group:   "hu",
			Service: "webgate-schema",
			Version: "0.0.0",
			Address: "",
		},
	}, nil)
	if err != nil {
		panic(err)
	}

	//cfg,err:=fd.ConfigFinder.UseAndSubscribeWithPrefix("schema",&configChangeHandler{})
	////fd.ConfigFinder.UseAndSubscribeConfig()
	//if err != nil{
	//	panic(err)
	//}
	//
	//fmt.Println(cfg)

	sss, err := fd.ServiceFinder.UseAndSubscribeService([]common.ServiceSubscribeItem{
		{
			ServiceName: "audio-moderation",
			ApiVersion:  "1.0.0",
		},
	}, &serviceHand{})
	if err != nil {
		panic(err)
	}

	fmt.Println(sss)
	select {}

}

type configChangeHandler struct {
}

func (c configChangeHandler) OnConfigFileChanged(config *common.Config) bool {
	fmt.Println("changed config", config)
	return true
}

func (c configChangeHandler) OnConfigFilesAdded(configs map[string]*common.Config) bool {
	fmt.Println("add config", configs)

	return true
}

func (c configChangeHandler) OnConfigFilesRemoved(configNames []string) bool {
	panic("implement me")
}

func (c configChangeHandler) OnError(errInfo common.ConfigErrInfo) {
	panic("implement me")
}

type serviceHand struct {
}

func (s serviceHand) OnServiceInstanceConfigChanged(name string, apiVersion string, addr string, config *common.ServiceInstanceConfig) bool {
	panic("implement me")
}

func (s serviceHand) OnServiceConfigChanged(name string, apiVersion string, config *common.ServiceConfig) bool {
	panic("implement me")
}

func (s serviceHand) OnServiceInstanceChanged(name string, apiVersion string, eventList []*common.ServiceInstanceChangedEvent) bool {
	panic("implement me")
}
