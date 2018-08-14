package main

import (
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

type TestConfig struct {
	CompanionUrl  string
	Address       string
	Project       string
	Group         string
	Service       string
	Version       string
	SubscribeFile []string
}

func main() {
	//	data, _ := fileutil.ReadFile("C:\\Users\\admin\\Desktop\\11.toml")
	//	fileutil.ParseTomlFile(data)
	//newProviderFinder("299.99.99.99:99")
	//newConsumerFinder("127.0.0.1:8082")
	// args := os.Args
	// if len(args) != 2 {
	// 	log.Println("参数错误")
	// 	return
	// }
	// file, _ := os.Open(args[1])
	// defer file.Close()
	// decoder := json.NewDecoder(file)
	// conf := TestConfig{}
	// err := decoder.Decode(&conf)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }
	//	fmt.Println(conf)
	//	newConfigFinder(conf)
	//newConfigFinder("127.0.0.1:10010", []string{"xsfs.toml"})
	//newProviderFinder("299.99.99.99:99")
	//newProviderFinder("299.99.99.99:100")
	//TODO  1. companion连不上怎么办，zk连不上怎么办
	newServiceFinder("299.99.99.99:101")
	for {
		time.Sleep(time.Second * 60)
		log.Println("I'm running.")
	}

}
func newServiceFinder(addr string) {

	cachePath, err := os.Getwd()
	if err != nil {
		return
	}
	cachePath += "/findercache"
	config := common.BootConfig{
		//CompanionUrl:     "http://companion.xfyun.iflytek:6868",
		CompanionUrl:  "http://10.1.87.70:6868",
		CachePath:     cachePath,
		ExpireTimeout: 5 * time.Second,
		MeteData: &common.ServiceMeteData{
			Project: "test0803",
			Group:   "test0803",
			Service: "test0803",
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
		//testGrayData(f)
		//testServiceAsync(f)
		testUseServiceAsync(f)
		//testUseService(f)

		//testConfigFeedback()
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
		CompanionUrl:  "http://10.1.87.70:6868",
		CachePath:     cachePath,
		ExpireTimeout: 5 * time.Second,
		MeteData: &common.ServiceMeteData{
			Project: "qq",
			Group:   "qq",
			Service: "qq",
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
		//testGrayData(f)
		//testServiceAsync(f)
		testRegisterService(f)
		//testUseService(f)

		//testConfigFeedback()
	}
}

func testRegisterService(f *finder.FinderManager) {
	f.ServiceFinder.RegisterService()

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

func newConfigFinder(conf TestConfig) {
	cachePath, err := os.Getwd()
	if err != nil {
		return
	}
	cachePath += "/findercache"
	config := common.BootConfig{
		//CompanionUrl:     "http://companion.xfyun.iflytek:6868",
		CompanionUrl:  conf.CompanionUrl,
		CachePath:     cachePath,
		ExpireTimeout: 5 * time.Second,
		MeteData: &common.ServiceMeteData{
			Project: conf.Project,
			Group:   conf.Group,
			Service: conf.Service,
			Version: conf.Version,
			Address: conf.Address,
		},
	}

	f, err := finder.NewFinderWithLogger(config, nil)
	if err != nil {
		fmt.Println(err)
	} else {
		testUseConfigAsyncByName(f, conf.SubscribeFile)

		//	testUserConfig(f, name)
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
		ServiceName:  "xrpc",
		ProviderList: []*common.ServiceInstance{},
		Config:       &common.ServiceConfig{},
	}
	instance := &common.ServiceInstance{
		Addr: "127.0.0.0:9091",
		Config: &common.ServiceInstanceConfig{
			IsValid: true,
		},
	}
	service.ProviderList = append(service.ProviderList, instance)

	err = finder.CacheService(cachepath, service)
	if err != nil {
		fmt.Println(err)
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
	//	handler := new(ServiceChangedHandle)
	item := []common.ServiceSubscribeItem{}
	item = append(item, common.ServiceSubscribeItem{ServiceName: "test0803", ApiVersion: "1.0"})
	serviceList, err := f.ServiceFinder.UseService(item)
	//serviceList, err := f.ServiceFinder.UseAndSubscribeService([]string{"iatExecutor"}, handler)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, s := range serviceList {
			fmt.Println(s.ServiceName, ":")
			for _, item := range s.ProviderList {
				fmt.Println("addr:", item.Addr)
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

func testGrayData(f *finder.FinderManager) {
	f.ConfigFinder.UseConfig([]string{"ddd"})
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

}

func testUseServiceAsync(f *finder.FinderManager) {
	handler := new(ServiceChangedHandle)
	var subscribe1 = common.ServiceSubscribeItem{ServiceName: "test0803", ApiVersion: "1.0"}
	var subscribe2 = common.ServiceSubscribeItem{ServiceName: "test0803", ApiVersion: "GGG3"}

	serviceList, err := f.ServiceFinder.UseAndSubscribeService([]common.ServiceSubscribeItem{subscribe1,subscribe2}, handler)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, s := range serviceList {
			fmt.Println("订阅的服务：",s.ServiceName, ":",s.ApiVersion ," --->")
			for _, item := range s.ProviderList {
				fmt.Println("----提供者地址 :")
				fmt.Println("--------:", item.Addr)
			}
		}
	}

}

func testUseConfigAsync(f *finder.FinderManager) {

	handler := ConfigChangedHandle{}
	count := 0

	f.InternalLogger.Info("The ", count, "th show:")
	//f.ConfigFinder.UseAndSubscribeConfig([]string{"test2.toml", "xsfc.toml.cfg"}, handler)
	configFiles, err := f.ConfigFinder.UseAndSubscribeConfig([]string{"2.yml"}, &handler)
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
func testUserConfig(f *finder.FinderManager, name []string) {
	configFiles, err := f.ConfigFinder.UseConfig(name)
	if err != nil {
		log.Println(err)
	}
	for _, c := range configFiles {
		log.Println(c.Name, ":\r\n", string(c.File))
	}

}

func testUseConfigAsyncByName(f *finder.FinderManager, name []string) {

	handler := ConfigChangedHandle{}
	//count := 0

	//f.InternalLogger.Info("The ", count, "th show:")
	//f.ConfigFinder.UseAndSubscribeConfig([]string{"test2.toml", "xsfc.toml.cfg"}, handler)
	configFiles, err := f.ConfigFinder.UseAndSubscribeConfig(name, &handler)
	if err != nil {
		log.Println(err)
	}
	for _, c := range configFiles {
		log.Println("首次获取配置文件名称：", c.Name, "  、\r\n内容为:\r\n", string(c.File))
	}
	f.ConfigFinder.UnSubscribeConfig("11.toml")
}