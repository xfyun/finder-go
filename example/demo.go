package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	finder "git.xfyun.cn/AIaaS/finder-go"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	"git.xfyun.cn/AIaaS/finder-go/utils/httputil"
)

func main() {
	newProviderFinder("127.0.0.1:8081")
	newConsumerFinder("127.0.0.1:8082")
	newConfigFinder("127.0.0.1:10010", []string{"xsfc.toml"})
	newConfigFinder("127.0.0.1:10010", []string{"xsfs.toml"})

	for {
		time.Sleep(time.Second * 60)
		log.Println("I'm running.")
	}

}

func newProviderFinder(addr string) {
	cachePath, err := os.Getwd()
	if err != nil {
		return
	}
	cachePath += "/findercache"
	config := common.BootConfig{
		//CompanionUrl:     "http://companion.xfyun.iflytek:6868",
		CompanionUrl:  "http://10.1.86.223:9080",
		CachePath:     cachePath,
		ExpireTimeout: 5 * time.Second,
		// MeteData: &common.ServiceMeteData{
		// 	Project: "project",
		// 	Group:   "group",
		// 	Service: "xsf",
		// 	Version: "1.0.0",
		// 	Address: "127.0.0.1:9091",
		// },
		// MeteData: &common.ServiceMeteData{
		// 	Project: "test",
		// 	Group:   "default",
		// 	Service: "xsf",
		// 	Version: "1.0.0",
		// 	Address: "127.0.0.1:9091",
		// },

		MeteData: &common.ServiceMeteData{
			Project: "AIaaS",
			Group:   "dx",
			Service: "finder_test",
			Version: "1.0",
			Address: addr,
		},

		// MeteData: &common.ServiceMeteData{
		// 	Project: "AIaaS",
		// 	Group:   "aitest",
		// 	Service: "atmos",
		// 	Version: "0.1",
		// 	Address: "127.0.0.1:8092",
		// },
	}

	f, err := finder.NewFinderWithLogger(config, nil)
	if err != nil {
		fmt.Println(err)
	} else {
		//testUseConfigAsync(f)
		//testCache(cachePath)
		testServiceAsync(f)
		//testUseService(f)

		//testConfigFeedback()
	}
}

func newConsumerFinder(addr string) {
	cachePath, err := os.Getwd()
	if err != nil {
		return
	}
	cachePath += "/findercache"
	config := common.BootConfig{
		//CompanionUrl:     "http://companion.xfyun.iflytek:6868",
		CompanionUrl:  "http://10.1.86.223:9080",
		CachePath:     cachePath,
		ExpireTimeout: 5 * time.Second,
		// MeteData: &common.ServiceMeteData{
		// 	Project: "project",
		// 	Group:   "group",
		// 	Service: "xsf",
		// 	Version: "1.0.0",
		// 	Address: "127.0.0.1:9091",
		// },
		// MeteData: &common.ServiceMeteData{
		// 	Project: "test",
		// 	Group:   "default",
		// 	Service: "xsf",
		// 	Version: "1.0.0",
		// 	Address: "127.0.0.1:9091",
		// },
		MeteData: &common.ServiceMeteData{
			Project: "AIaaS",
			Group:   "dx",
			Service: "finder_test",
			Version: "1.0",
			Address: addr,
		},
	}

	f, err := finder.NewFinderWithLogger(config, nil)
	if err != nil {
		fmt.Println(err)
	} else {
		//testUseConfigAsync(f)
		//testCache(cachePath)
		testUseServiceAsync(f)

		//testConfigFeedback()
	}
}

func newConfigFinder(addr string, name []string) {
	cachePath, err := os.Getwd()
	if err != nil {
		return
	}
	cachePath += "/findercache"
	config := common.BootConfig{
		//CompanionUrl:     "http://companion.xfyun.iflytek:6868",
		CompanionUrl:  "http://10.1.86.223:9080",
		CachePath:     cachePath,
		ExpireTimeout: 5 * time.Second,
		MeteData: &common.ServiceMeteData{
			Project: "AIaaS",
			Group:   "dx",
			Service: "iatExecutor",
			Version: "2.0.7",
			Address: addr,
		},
	}

	f, err := finder.NewFinderWithLogger(config, nil)
	if err != nil {
		fmt.Println(err)
	} else {
		testUseConfigAsyncByName(f, name)
		//testCache(cachePath)
		//testUseServiceAsync(f)

		//testConfigFeedback()
	}
}

func getLocalIP(url string) (string, error) {
	var host string
	var port string
	var localIP string
	items := strings.Split(url, ":")
	if len(items) == 3 {
		host = strings.Replace(items[1], "/", "", -1)
		port = items[2]
	} else if len(items) == 2 {
		host = strings.Replace(items[0], "/", "", -1)
		port = items[1]
	} else {
		host = url
		port = "80"
	}

	if len(host) == 0 {
		return "", errors.New("testRemote:invalid remote url")
	}
	if len(port) == 0 {
		port = "80"
	}
	ips, err := net.LookupHost(host)
	if err != nil {
		return "", err
	}
	for _, ip := range ips {
		conn, err := net.Dial("tcp", ip+":"+port)
		if err != nil {
			log.Println("testRemote:", err)
			continue
		}
		localIP = conn.LocalAddr().String()
		log.Println("testRemote:ok")
		err = conn.Close()
		if err != nil {
			log.Println("testRemote:", err)
			break
		}
		break
	}

	if len(localIP) == 0 {
		return "", errors.New("testRemote:failed")
	}

	fmt.Println("local ip:", localIP)

	return localIP, nil
}

func testCache(cachepath string) {
	configFile := `[test]\r\n\titem = "value"`
	config := &common.Config{
		Name: "default.cfg",
		File: []byte(configFile),
	}
	err := finder.CacheConfig(cachepath, config)
	if err != nil {
		fmt.Println(err)
	}
	c, err := finder.GetConfigFromCache(cachepath, "default.cfg")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("default.cfg:", string(c.File))
	}

	zkInfo := &common.StorageInfo{
		Addr:            []string{"10.1.86.73:2181", "10.1.86.74:2181"},
		ConfigRootPath:  "/polaris/config/",
		ServiceRootPath: "/polaris/service/",
	}
	err = finder.CacheStorageInfo(cachepath, zkInfo)
	if err != nil {
		fmt.Println(err)
	}
	newZkInfo, err := finder.GetStorageInfoFromCache(cachepath)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("ZkAddr:", newZkInfo.Addr)
		fmt.Println("ConfigRootPath:", newZkInfo.ConfigRootPath)
		fmt.Println("ServiceRootPath:", newZkInfo.ServiceRootPath)
	}

	service := &common.Service{
		Name:       "xrpc",
		ServerList: []*common.ServiceInstance{},
		Config: &common.ServiceConfig{
			ProxyMode:       "default",
			LoadBalanceMode: "default",
		},
	}
	instance := &common.ServiceInstance{
		Addr: "127.0.0.0:9091",
		Config: &common.ServiceInstanceConfig{
			Weight:  100,
			IsValid: true,
		},
	}
	service.ServerList = append(service.ServerList, instance)

	err = finder.CacheService(cachepath, service)
	if err != nil {
		fmt.Println(err)
	}
	newService, err := finder.GetServiceFromCache(cachepath, "xrpc")
	if err != nil {
		fmt.Println(err)
	} else {
		data, err := json.Marshal(newService)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("newService", string(data))
		}
	}
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

func testUseService(f *finder.FinderManager) {
	handler := new(ServiceChangedHandle)
	serviceList, err := f.ServiceFinder.UseAndSubscribeService([]string{"iatExecutor"}, handler)
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

func testServiceAsync(f *finder.FinderManager) {
	var err error
	err = f.ServiceFinder.RegisterService()
	//err = f.ServiceFinder.RegisterServiceWithAddr("10.1.203.36:50052")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("RegisterService is ok.")
	}
	time.Sleep(time.Second * 2)
	return

	// serviceList, err := f.ServiceFinder.UseService([]string{"xrpc"})
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	for _, s := range serviceList {
	// 		fmt.Println(s.Name, ":")
	// 		for _, item := range s.ServerList {
	// 			fmt.Println("addr:", item.Addr)
	// 			fmt.Println("weight:", item.Config.Weight)
	// 			fmt.Println("is_valid:", item.Config.IsValid)
	// 		}
	// 	}

	// 	time.Sleep(time.Second * 2)
	// }

	forIndex := 0
	for {
		forIndex++
		go func() {

			handler := new(ServiceChangedHandle)
			fmt.Println("use ", forIndex)
			serviceList, err := f.ServiceFinder.UseAndSubscribeService([]string{"aitest_weiwang26", "aitest_weiwang26"}, handler)
			fmt.Println("use end ", forIndex)
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

				//time.Sleep(time.Second * 2)
			}
		}()

		if forIndex > 10 {
			break
		}
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

func testUseServiceAsync(f *finder.FinderManager) {
	handler := new(ServiceChangedHandle)
	serviceList, err := f.ServiceFinder.UseAndSubscribeService([]string{"iatExecutor"}, handler)
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
	}

	return

	forIndex := 0
	for {
		forIndex++
		go func() {

			handler := new(ServiceChangedHandle)
			fmt.Println("use ", forIndex)
			serviceList, err := f.ServiceFinder.UseAndSubscribeService([]string{"aitest_weiwang26", "aitest_weiwang26"}, handler)
			fmt.Println("use end ", forIndex)
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

				//time.Sleep(time.Second * 2)
			}
		}()

		if forIndex > 10 {
			break
		}
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
	count := 0

	f.InternalLogger.Info("The ", count, "th show:")
	//f.ConfigFinder.UseAndSubscribeConfig([]string{"test2.toml", "xsfc.toml.cfg"}, handler)
	configFiles, err := f.ConfigFinder.UseAndSubscribeConfig([]string{"2.yml"}, handler)
	if err != nil {
		log.Println(err)
	}
	for _, c := range configFiles {
		log.Println(c.Name, ":\r\n", string(c.File))
	}

	for {
		//fmt.Println("The ", count, "th show:")
		//configFiles, err := f.ConfigFinder.UseAndSubscribeConfig([]string{"test2.toml", "xsfc.tmol"}, handler)

		//f.ConfigFinder.UseAndSubscribeConfig([]string{"test2.toml", "xsfc.tmol"}, handler)
		//configFiles, err := f.ConfigFinder.UseConfig([]string{"xsfc.tmol"})

		if count > 200 {
			f.ConfigFinder.UnSubscribeConfig("default.toml")
		}
		if count > 600 {
			break
		}
		count++
		time.Sleep(time.Second * 1)
	}

}

func testUseConfigAsyncByName(f *finder.FinderManager, name []string) {
	// configFiles, err := f.ConfigFinder.UseConfig([]string{"test.toml"})
	// if err != nil {
	// 	fmt.Println(err)
	// }
	handler := new(ConfigChangedHandle)
	count := 0

	f.InternalLogger.Info("The ", count, "th show:")
	//f.ConfigFinder.UseAndSubscribeConfig([]string{"test2.toml", "xsfc.toml.cfg"}, handler)
	configFiles, err := f.ConfigFinder.UseAndSubscribeConfig(name, handler)
	if err != nil {
		log.Println(err)
	}
	for _, c := range configFiles {
		log.Println(c.Name, ":\r\n", string(c.File))
	}

	return
	for {
		//fmt.Println("The ", count, "th show:")
		//configFiles, err := f.ConfigFinder.UseAndSubscribeConfig([]string{"test2.toml", "xsfc.tmol"}, handler)

		//f.ConfigFinder.UseAndSubscribeConfig([]string{"test2.toml", "xsfc.tmol"}, handler)
		//configFiles, err := f.ConfigFinder.UseConfig([]string{"xsfc.tmol"})

		if count > 200 {
			f.ConfigFinder.UnSubscribeConfig("default.toml")
		}
		if count > 600 {
			break
		}
		count++
		time.Sleep(time.Second * 1)
	}

}
