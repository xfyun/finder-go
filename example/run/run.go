package main

import (
	"fmt"
	"git.iflytek.com/AIaaS/finder-go"
	common "git.iflytek.com/AIaaS/finder-go/common"
	"time"
)

func main(){
	fd,err:=finder.NewFinderWithLogger(common.BootConfig{
		CompanionUrl:  "http://10.1.87.69:6868",
		CachePath:     "./findercache",
		CacheConfig:   false,
		CacheService:  false,
		ExpireTimeout: 5*time.Second,
		MeteData:    & common.ServiceMeteData{
			Project: "AIPaaS",
			Group:   "hu",
			Service: "webgate-schema",
			Version: "0.0.0",
			Address: "",
		} ,
	},nil)
	if err != nil{
		panic(err)
	}

	cfg,err:=fd.ConfigFinder.UseAndSubscribeWithPrefix("schema",&configChangeHandler{})
	//fd.ConfigFinder.UseAndSubscribeConfig()
	if err != nil{
		panic(err)
	}

	fmt.Println(cfg)
	select {
	}

}


type configChangeHandler struct {

}

func (c configChangeHandler) OnConfigFileChanged(config *common.Config) bool {
	fmt.Println("changed config",config)
	return true
}

func (c configChangeHandler) OnConfigFilesAdded(configs map[string]*common.Config) bool {
	fmt.Println("add config",configs)

	return true
}

func (c configChangeHandler) OnConfigFilesRemoved(configNames []string) bool {
	panic("implement me")
}

func (c configChangeHandler) OnError(errInfo common.ConfigErrInfo) {
	panic("implement me")
}
