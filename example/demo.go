package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
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
	args := os.Args
	if len(args) != 2 {
		log.Println("参数错误")
		return
	}
	file, _ := os.Open(args[1])
	defer file.Close()
	decoder := json.NewDecoder(file)
	conf := TestConfig{}
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	//	fmt.Println(conf)

	newConfigFinder(conf)
	//newConfigFinder("127.0.0.1:10010", []string{"xsfs.toml"})

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
		CompanionUrl:  "http://10.1.87.70:6868",
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
		testGrayData(f)
		//testServiceAsync(f)
		//testUseService(f)

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

func testGrayData(f *finder.FinderManager) {
	f.ConfigFinder.UseConfig([]string{"ddd"})
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
	// configFiles, err := f.ConfigFinder.UseConfig([]string{"test.toml"})
	// if err != nil {
	// 	fmt.Println(err)
	// }
	handler := new(ConfigChangedHandle)
	//count := 0

	//f.InternalLogger.Info("The ", count, "th show:")
	//f.ConfigFinder.UseAndSubscribeConfig([]string{"test2.toml", "xsfc.toml.cfg"}, handler)
	configFiles, err := f.ConfigFinder.UseAndSubscribeConfig(name, handler)
	if err != nil {
		log.Println(err)
	}
	for _, c := range configFiles {
		log.Println("首次获取配置文件名称：", c.Name, "  、\r\n内容为:\r\n", string(c.File))
	}
	//f.ConfigFinder.UnSubscribeConfig("11.toml")
}
