package finder

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	companion "git.xfyun.cn/AIaaS/finder-go/companion"
	"git.xfyun.cn/AIaaS/finder-go/storage"
	"git.xfyun.cn/AIaaS/finder-go/utils/fileutil"
)

const (
	SERVICE_INSTANCE_CHANGED        = "SERVICE_INSTANCE"
	SERVICE_CONFIG_CHANGED          = "SERVICE_CONFIG"
	SERVICE_INSTANCE_CONFIG_CHANGED = "SERVICE_INSTANCE_CONFIG"
	CONFIG_CHANGED                  = "CONFIG"
	GRAY_CONFIG_CHANGED             = "GRAY_CONFIG"
)

type ServiceChangedCallback struct {
	name      string
	eventType string
	uh        common.ServiceChangedHandler
	bootCfg   *common.BootConfig
	sm        storage.StorageManager
	root      string
}

func NewServiceChangedCallback(serviceName string, watchType string, rootPath string, userHandle common.ServiceChangedHandler, bootConfig *common.BootConfig, storageMgr storage.StorageManager) ServiceChangedCallback {
	return ServiceChangedCallback{
		name:      serviceName,
		eventType: watchType,
		root:      rootPath,
		uh:        userHandle,
		bootCfg:   bootConfig,
		sm:        storageMgr,
	}
}

// func (cb *ServiceChangedCallback) checkEventType(name string, path string) bool {
// 	if paths, ok := cb.watchedTypes.Load(name); ok {
// 		return arrayutil.Contains(path, paths)
// 	}

// 	return false
// }

func (cb *ServiceChangedCallback) DataChangedCallback(path string, node string, data []byte) {
	if cb.eventType == SERVICE_CONFIG_CHANGED {
		cb.OnServiceConfigChanged(cb.name, data)
	} else if cb.eventType == SERVICE_INSTANCE_CONFIG_CHANGED {
		cb.OnServiceInstanceConfigChanged(cb.name, node, data)
	}
}

func (cb *ServiceChangedCallback) ChildrenChangedCallback(path string, node string, children []string) {
	if cb.eventType == SERVICE_INSTANCE_CHANGED {
		cb.OnServiceInstanceChanged(cb.name, children)
	}
}

func (cb *ServiceChangedCallback) OnServiceInstanceConfigChanged(name string, addr string, data []byte) {
	pushID, config, err := common.DecodeValue(data)
	if err != nil {
		// todo
		return
	}

	f := &common.ServiceFeedback{
		PushID:          pushID,
		ServiceMete:     cb.bootCfg.MeteData,
		Provider:        name,
		ProviderVersion: "",
		UpdateTime:      time.Now().Unix(),
		UpdateStatus:    1,
	}
	c := &common.ServiceInstanceConfig{}
	err = json.Unmarshal(config, c)
	if err != nil {
		f.LoadStatus = -1
		log.Println(err)
	} else {
		ok := cb.uh.OnServiceInstanceConfigChanged(name, addr, c)
		if ok {
			log.Println("load success:", pushID)
			f.LoadStatus = 1
		}
	}

	f.LoadTime = time.Now().Unix()
	err = pushServiceFeedback(cb.bootCfg.CompanionUrl, f)
	if err != nil {
		log.Println(err)
	}
}

func (cb *ServiceChangedCallback) OnServiceConfigChanged(name string, data []byte) {
	pushID, config, err := common.DecodeValue(data)
	if err != nil {
		// todo
		return
	}

	f := &common.ServiceFeedback{
		PushID:          pushID,
		ServiceMete:     cb.bootCfg.MeteData,
		Provider:        name,
		ProviderVersion: "",
		UpdateTime:      time.Now().Unix(),
		UpdateStatus:    1,
	}
	c := &common.ServiceConfig{}
	err = json.Unmarshal(config, c)
	if err != nil {
		f.LoadStatus = -1
		log.Println(err)
	} else {
		ok := cb.uh.OnServiceConfigChanged(name, c)
		if ok {
			log.Println("load success:", pushID)
			f.LoadStatus = 1
		}
	}

	f.LoadTime = time.Now().Unix()
	err = pushServiceFeedback(cb.bootCfg.CompanionUrl, f)
	if err != nil {
		log.Println(err)
	}
}

func (cb *ServiceChangedCallback) OnServiceInstanceChanged(name string, addrList []string) {
	eventList := make([]*common.ServiceInstanceChangedEvent, 0)
	newInstances := []*common.ServiceInstance{}
	cachedService, err := GetServiceFromCache(cb.bootCfg.CachePath, name)
	if err != nil {
		log.Println("GetServiceFromCache", name, err)
		cachedService = &common.Service{Name: name, ServerList: newInstances}
	}
	if len(addrList) > 0 {
		servicePath := fmt.Sprintf("%s/%s/provider", cb.root, name)
		if len(cachedService.ServerList) > 0 {
			oldInstances, deletedEvent := getDeletedInstEvent(addrList, cachedService.ServerList)
			if deletedEvent != nil {
				eventList = append(eventList, deletedEvent)
			}
			if oldInstances != nil {
				newInstances = append(newInstances, oldInstances...)
			}
			addedEvent := getAddedInstEvents(cb.sm, servicePath, addrList, cachedService.ServerList)
			if addedEvent != nil {
				newInstances = append(newInstances, addedEvent.ServerList...)
				eventList = append(eventList, addedEvent)
			}
		} else {
			addedEvent := getAddedInstEvents(cb.sm, servicePath, addrList, cachedService.ServerList)
			if addedEvent != nil {
				newInstances = append(newInstances, addedEvent.ServerList...)
				eventList = append(eventList, addedEvent)
			}
		}
	} else {
		oldInstances, deletedEvent := getDeletedInstEvent(addrList, cachedService.ServerList)
		if deletedEvent != nil {
			eventList = append(eventList, deletedEvent)
		}
		if oldInstances != nil {
			newInstances = append(newInstances, oldInstances...)
		}
	}

	cachedService.ServerList = newInstances
	err = CacheService(cb.bootCfg.CachePath, cachedService)
	if err != nil {
		log.Println("CacheService failed")
	}

	ok := cb.uh.OnServiceInstanceChanged(name, eventList)
	if !ok {
		log.Println("OnServiceInstanceChanged is not ok")
	}
}

func getDeletedInstEvent(addrList []string, insts []*common.ServiceInstance) ([]*common.ServiceInstance, *common.ServiceInstanceChangedEvent) {
	var event *common.ServiceInstanceChangedEvent
	var oldInstances []*common.ServiceInstance
	var deletedInstances []*common.ServiceInstance
	var deleted bool
	for _, inst := range insts {
		deleted = true
		for _, addr := range addrList {
			if addr == inst.Addr {
				deleted = false
				if oldInstances == nil {
					oldInstances = []*common.ServiceInstance{}
				}
				oldInstances = append(oldInstances, inst)
			}
		}
		if deleted {
			if deletedInstances == nil {
				deletedInstances = []*common.ServiceInstance{}
			}
			deletedInstances = append(deletedInstances, inst)
		}
	}

	if deletedInstances != nil {
		event = &common.ServiceInstanceChangedEvent{
			EventType:  common.INSTANCEREMOVE,
			ServerList: deletedInstances,
		}
	}

	return oldInstances, event
}

func getAddedInstEvents(sm storage.StorageManager, servicePath string, addrList []string, insts []*common.ServiceInstance) *common.ServiceInstanceChangedEvent {
	var event *common.ServiceInstanceChangedEvent
	var addedInstances []*common.ServiceInstance
	var added bool
	for _, addr := range addrList {
		added = true
		for _, inst := range insts {
			if addr == inst.Addr {
				added = false
			}
		}
		if added {
			inst, err := getServiceInstance(sm, servicePath, addr)
			if err != nil {
				log.Println(err)
				// todo
				continue
			}

			if addedInstances == nil {
				addedInstances = []*common.ServiceInstance{}
			}
			addedInstances = append(addedInstances, inst)
		}
	}

	if addedInstances != nil {
		event = &common.ServiceInstanceChangedEvent{
			EventType:  common.INSTANCEADDED,
			ServerList: addedInstances,
		}
	}

	return event
}

type ConfigChangedCallback struct {
	name        string
	eventType   string
	grayGroupId string
	uh          common.ConfigChangedHandler
	bootCfg     *common.BootConfig
	sm          storage.StorageManager
	root        string
}

func NewConfigChangedCallback(serviceName string, watchType string, rootPath string, userHandle common.ConfigChangedHandler, bootConfig *common.BootConfig, storageMgr storage.StorageManager) ConfigChangedCallback {
	return ConfigChangedCallback{
		name:      serviceName,
		eventType: watchType,
		root:      rootPath,
		uh:        userHandle,
		bootCfg:   bootConfig,
		sm:        storageMgr,
	}
}

func (cb *ConfigChangedCallback) DataChangedCallback(path string, node string, data []byte) {
	if cb.eventType == CONFIG_CHANGED {
		cb.OnConfigFileChanged(cb.name, data)
	} else if cb.eventType == GRAY_CONFIG_CHANGED {
		cb.OnGrayConfigChanged(cb.name, data)
	}

}

func (cb *ConfigChangedCallback) ChildrenChangedCallback(path string, node string, children []string) {

}

func (cb *ConfigChangedCallback) OnGrayConfigChanged(name string, data []byte) {

	groupId := cb.grayGroupId
	if grayGroupId, ok := ParseGrayConfigData(cb.bootCfg.MeteData.Address, data); ok {
		if strings.Compare(groupId, grayGroupId) == 0 {
			//如果之前的group和现在的一样，则代表没有切换灰度组。直接结束
			return
		} else {
			//不相等，则代表灰度组有改变。需要重新获取节点配置信息
			basePath := cb.root + "/gray/" + grayGroupId + "/" + cb.name
			data, err := cb.sm.GetDataWithWatch(basePath, cb)
			if err != nil {
				logger.Info(" [OnGrayConfigChanged] 重新从路径 ", basePath, " 获取灰度配置失败 ", err)
				return
			}
			//成功的话，调用OnConfigFileChanged来执行用户设定的回调函数
			cb.OnConfigFileChanged(cb.name, data)
		}

	} else if len(groupId) != 0 {
		//现在不在灰度组，以前在灰度组中，则需要重新获取配置信息
		data, err := cb.sm.GetDataWithWatch(cb.root+"/"+cb.name, cb)
		if err != nil {
			logger.Info(" [OnGrayConfigChanged] 重新从路径 ", cb.root+"/"+cb.name, " 获取灰度配置失败 ", err)
			return
		}
		cb.OnConfigFileChanged(cb.name, data)
	} else {
		//当前不在灰度组中，以前也不在，则配置信息无需再次更新
		return
	}
}
func (cb *ConfigChangedCallback) OnConfigFileChanged(name string, data []byte) {
	pushID, file, err := common.DecodeValue(data)
	if err != nil {
		// todo
	} else {
		f := &common.ConfigFeedback{
			PushID:       pushID,
			ServiceMete:  cb.bootCfg.MeteData,
			Config:       name,
			UpdateTime:   time.Now().Unix(),
			UpdateStatus: 1,
		}
		tomlConfig := make(map[string]interface{})
		if fileutil.IsTomlFile(name) {
			tomlConfig = fileutil.ParseTomlFile(file)
		}
		c := &common.Config{
			Name:      name,
			File:      file,
			ConfigMap: tomlConfig,
		}

		ok := cb.uh.OnConfigFileChanged(c)
		if ok {
			err = CacheConfig(cb.bootCfg.CachePath, c)
			if err != nil {
				log.Println(err)
				// todo
			}

			log.Println("load success:", pushID)
			f.LoadStatus = 1
		}
		f.LoadTime = time.Now().Unix()
		err = pushConfigFeedback(cb.bootCfg.CompanionUrl, f)
		if err != nil {
			log.Println(err)
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

func pushService(companionUrl string, project string, group string, service string) error {
	url := companionUrl + "/finder/register_service_info"
	return companion.RegisterService(hc, url, project, group, service)
}
