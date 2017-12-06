package main

import (
	"finder-go"
	"finder-go/common"
	"fmt"
	"os"
	"time"
)

func main() {
	cachePath, err := os.Getwd()
	if err != nil {
		return
	}
	cachePath += "/findercache"
	config := common.BootConfig{
		CompanionUrl:     "http://127.0.0.1:9090",
		CachePath:        cachePath,
		TickerDuration:   5000,
		ZkSessionTimeout: 30 * time.Second,
		ZkConnectTimeout: 3 * time.Second,
		ZkMaxSleepTime:   15 * time.Second,
		ZkMaxRetryNum:    3,
		MeteData: common.ServiceMeteData{
			Project: "test",
			Group:   "default",
			Service: "xrpc",
			Version: "1.0.0",
		},
	}

	f, err := finder.NewFinder(config)
	if err != nil {
		fmt.Println(err)
	}
	// configList, err := finder.ConfigFinder.UseConfig([]string{"test.toml"})
	// if err != nil {
	// 	fmt.Println(err)
	// }

	configFiles, err := f.ConfigFinder.UseAndSubscribeConfig([]string{"test.toml", "default.toml"}, func(c common.Config) {
		fmt.Println(c.Name, " has changed:\r\n", string(c.File))
	})

	if err != nil {
		fmt.Println(err)
	}
	for _, c := range configFiles {
		fmt.Println(c.Name, ":\r\n", string(c.File))
	}

	count := 0
	for {
		count++
		if count > 20 {
			f.ConfigFinder.UnSubscribeConfig("default.toml")
		}
		if count > 60 {
			break
		}
		time.Sleep(time.Second * 1)
	}

}
