package finder

import (
	"fmt"
	"git.iflytek.com/AIaaS/finder-go/common"
	"testing"
	"time"
)

func TestTT(t *testing.T) {
	fd ,err := NewFinderWithLogger(common.BootConfig{
		CompanionUrl:  "http://10.1.87.69:6868",
		CachePath:     "./fdcache",
		CacheConfig:   false,
		CacheService:  false,
		ExpireTimeout: 5*time.Second,
		MeteData:      &common.ServiceMeteData{
			Project: "AIPaaS",
			Group:   "hu",
			Service: "webgate-schema",
			Version: "0.0.0",
			Address: "",
		},
	},nil)

	if err != nil{
		panic(err)
	}

	cfgs ,err := fd.ConfigFinder.UseAndSubscribeConfig([]string{"schema_activeaudiofea.json"},&configFinder{})
	if err != nil{
		panic(err)
	}
	fmt.Println(cfgs)

	cfgs ,err = fd.ConfigFinder.UseAndSubscribeWithPrefix("/",&configFinder{})
	if err != nil{
		panic(err)
	}
	fmt.Println(cfgs)

}



type configFinder struct {

}

func (c configFinder) OnConfigFileChanged(config *common.Config) bool {
	panic("implement me")
}

func (c configFinder) OnConfigFilesAdded(configs map[string]*common.Config) bool {
	panic("implement me")
}

func (c configFinder) OnConfigFilesRemoved(configNames []string) bool {
	panic("implement me")
}

func (c configFinder) OnError(errInfo common.ConfigErrInfo) {
	panic("implement me")
}
