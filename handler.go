package finder

import (
	"encoding/json"
	"finder-go/common"
	"finder-go/companion"
	"finder-go/utils/zkutil"
	"fmt"
	"time"
)

type ServiceHandle struct {
	ChangedHandler common.ServiceChangedHandler
	config         *common.BootConfig
	zkManager      *zkutil.ZkManager
}

func (s *ServiceHandle) OnServiceInstanceConfigChanged(name string, addr string, data []byte) {
	pushID, config, err := common.DecodeValue(data)
	if err != nil {
		// todo
		return
	}

	f := &common.ServiceFeedback{
		PushID:          pushID,
		ServiceMete:     s.config.MeteData,
		Provider:        name,
		ProviderVersion: "",
		UpdateTime:      time.Now().Unix(),
		UpdateStatus:    1,
	}
	c := &common.ServiceInstanceConfig{}
	err = json.Unmarshal(config, c)
	if err != nil {
		f.LoadStatus = -1
		fmt.Println(err)
	} else {
		ok := s.ChangedHandler.OnServiceInstanceConfigChanged(name, addr, c)
		if ok {
			fmt.Println("load success:", pushID)
			f.LoadStatus = 1
		}
	}

	f.LoadTime = time.Now().Unix()
	err = pushServiceFeedback(s.config.CompanionUrl, f)
	if err != nil {
		fmt.Println(err)
	}
}

func (s *ServiceHandle) OnServiceConfigChanged(name string, data []byte) {
	pushID, config, err := common.DecodeValue(data)
	if err != nil {
		// todo
		return
	}

	f := &common.ServiceFeedback{
		PushID:          pushID,
		ServiceMete:     s.config.MeteData,
		Provider:        name,
		ProviderVersion: "",
		UpdateTime:      time.Now().Unix(),
		UpdateStatus:    1,
	}
	c := &common.ServiceConfig{}
	err = json.Unmarshal(config, c)
	if err != nil {
		f.LoadStatus = -1
		fmt.Println(err)
	} else {
		ok := s.ChangedHandler.OnServiceConfigChanged(name, c)
		if ok {
			fmt.Println("load success:", pushID)
			f.LoadStatus = 1
		}
	}

	f.LoadTime = time.Now().Unix()
	err = pushServiceFeedback(s.config.CompanionUrl, f)
	if err != nil {
		fmt.Println(err)
	}
}

func (s *ServiceHandle) OnServiceInstanceChanged(name string, addrList []string) {
	instances := make([]*common.ServiceInstance, 0)
	if len(addrList) > 0 {
		servicePath := fmt.Sprintf("%s/%s/provider", s.zkManager.MetaData.ServiceRootPath, s.config.MeteData.Service)
		for _, inst := range addrList {
			serviceInstance, err := getServiceInstance(s.zkManager, servicePath, inst)
			if err != nil {
				fmt.Println(err)
				// todo
				continue
			}

			instances = append(instances, serviceInstance)
		}
	}
	s.ChangedHandler.OnServiceInstanceChanged(name, instances)
}

type ConfigHandle struct {
	config         *common.BootConfig
	ChangedHandler common.ConfigChangedHandler
}

func (s *ConfigHandle) OnConfigFileChanged(name string, data []byte) {
	pushID, file, err := common.DecodeValue(data)
	if err != nil {
		// todo
	} else {
		f := &common.ConfigFeedback{
			PushID:       pushID,
			ServiceMete:  s.config.MeteData,
			Config:       name,
			UpdateTime:   time.Now().Unix(),
			UpdateStatus: 1,
		}
		c := &common.Config{
			Name: name,
			File: file,
		}

		ok := s.ChangedHandler.OnConfigFileChanged(c)
		if ok {
			fmt.Println("load success:", pushID)
			f.LoadStatus = 1
		}
		f.LoadTime = time.Now().Unix()
		err = pushConfigFeedback(s.config.CompanionUrl, f)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func pushConfigFeedback(companionUrl string, f *common.ConfigFeedback) error {
	url := companionUrl + "/finder/push_config_feedback"
	return companion.FeedbackForConfig(hc, url, f)
}

func pushServiceFeedback(companionUrl string, f *common.ServiceFeedback) error {
	url := companionUrl + "/finder/push_service_feedback"
	return companion.FeedbackForService(hc, url, f)
}
