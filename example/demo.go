package main

import (
	"finder-go"
	"finder-go/common"
	"finder-go/utils/httputil"
	"fmt"
	"net"
	"net/http"
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
		CompanionUrl:     "http://10.1.86.223:9080",
		CachePath:        cachePath,
		TickerDuration:   5000,
		ZkSessionTimeout: 3 * time.Second,
		ZkConnectTimeout: 300 * time.Second,
		ZkMaxSleepTime:   15 * time.Second,
		ZkMaxRetryNum:    3,
		MeteData: &common.ServiceMeteData{
			Project: "test",
			Group:   "default",
			Service: "xrpc",
			Version: "1.0.0",
			Address: "127.0.0.1:9091",
		},
	}

	f, err := finder.NewFinder(config)
	if err != nil {
		fmt.Println(err)
	}

	//testUseConfigAsync(f)
	testServiceAsync(f)

	//testConfigFeedback()

}

func testConfigFeedback() {
	url := "http://10.1.200.75:9080/finder/push_config_feedback"
	contentType := "application/x-www-form-urlencoded"
	hc := &http.Client{
		Transport: &http.Transport{
			Dial: func(nw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(1 * time.Second)
				c, err := net.DialTimeout(nw, addr, time.Second*1)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	params := []byte("pushId=123456&project=test&group=default&service=xrpc&version=1.0.0&config=default.cfg&addr=10.1.86.221:9091&update_status=1&update_time=1513044755&load_status=1&load_time=1513044757")
	result, err := httputil.DoPost(hc, contentType, url, params)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(result))
}

func testServiceAsync(f *finder.FinderManager) {
	var err error
	err = f.ServiceFinder.RegisterService()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("RegisterService is ok.")
	}
	time.Sleep(time.Second * 2)

	var serviceList []*common.Service
	serviceList, err = f.ServiceFinder.UseService([]string{"xrpc"})
	if err != nil {
		fmt.Println(err)
	} else {
		for _, s := range serviceList {
			fmt.Println(s.Name, ":")
			for _, item := range s.ServerList {
				fmt.Println("addr:", item.Addr)
				fmt.Println("weight:", item.Config.Weight)
				fmt.Println("is_valid:", item.Config.IsValid)
			}
		}

		time.Sleep(time.Second * 2)
	}

	handler := new(ServiceChangedHandle)
	serviceList, err = f.ServiceFinder.UseAndSubscribeService([]string{"xrpc"}, handler)

	if err != nil {
		fmt.Println(err)
	} else {
		for _, s := range serviceList {
			fmt.Println(s.Name, ":")
			for _, item := range s.ServerList {
				fmt.Println("addr:", item.Addr)
				fmt.Println("weight:", item.Config.Weight)
				fmt.Println("is_valid:", item.Config.IsValid)
			}
		}

		time.Sleep(time.Second * 2)
	}

	count := 0
	for {
		count++
		if count > 200 {
			//f.ConfigFinder.UnSubscribeConfig("default.toml")
		}
		if count > 600 {
			err = f.ServiceFinder.UnRegisterService()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("UnRegisterService is ok.")
			}
			break
		}
		time.Sleep(time.Second * 1)
	}
}

func testUseConfigAsync(f *finder.FinderManager) {

	// configFiles, err := f.ConfigFinder.UseConfig([]string{"test.toml"})
	// if err != nil {
	// 	fmt.Println(err)
	// }
	handler := new(ConfigChangedHandle)
	configFiles, err := f.ConfigFinder.UseAndSubscribeConfig([]string{"test2.toml", "xsfc.tmol"}, handler)

	if err != nil {
		fmt.Println(err)
	}
	for _, c := range configFiles {
		fmt.Println(c.Name, ":\r\n", string(c.File))
	}

	count := 0
	for {
		count++
		if count > 200 {
			f.ConfigFinder.UnSubscribeConfig("default.toml")
		}
		if count > 600 {
			break
		}
		time.Sleep(time.Second * 1)
	}
}
