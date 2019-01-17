package finder

import (
	"time"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	companion "git.xfyun.cn/AIaaS/finder-go/companion"
	"git.xfyun.cn/AIaaS/finder-go/log"
	"git.xfyun.cn/AIaaS/finder-go/route"
	"git.xfyun.cn/AIaaS/finder-go/storage"
	"git.xfyun.cn/AIaaS/finder-go/utils/fileutil"
	"git.xfyun.cn/AIaaS/finder-go/utils/serviceutil"
	"strings"
)

const (
	SERVICE_INSTANCE_CHANGED        = "SERVICE_INSTANCE"
	SERVICE_CONFIG_CHANGED          = "SERVICE_CONFIG"
	SERVICE_INSTANCE_CONFIG_CHANGED = "SERVICE_INSTANCE_CONFIG"
	SERVICE_ROUTE_CHANGED           = "SERVICE_ROUTE"
	CONFIG_CHANGED                  = "CONFIG"
	GRAY_CONFIG_CHANGED             = "GRAY_CONFIG"
)

type ServiceChangedCallback struct {
	serviceItem   common.ServiceSubscribeItem
	eventType     string
	uh            common.ServiceChangedHandler
	serviceFinder *ServiceFinder
}

func NewServiceChangedCallback(serviceItem common.ServiceSubscribeItem, watchType string, serviceFinder *ServiceFinder, userHandle common.ServiceChangedHandler) ServiceChangedCallback {
	return ServiceChangedCallback{
		serviceItem:   serviceItem,
		eventType:     watchType,
		uh:            userHandle,
		serviceFinder: serviceFinder,
	}
}

func (cb *ServiceChangedCallback) DataChangedCallback(path string, node string, data []byte) {
	cb.serviceFinder.locker.Lock()
	defer cb.serviceFinder.locker.Unlock()
	log.Log.Infof("recv callback ,path : %v ,eventType: %v", path, cb.eventType)
	if cb.eventType == SERVICE_CONFIG_CHANGED {
		cb.OnServiceConfigChanged(cb.serviceItem, data)
	} else if cb.eventType == SERVICE_INSTANCE_CONFIG_CHANGED {
		cb.OnServiceInstanceConfigChanged(cb.serviceItem, node, data)
	} else if cb.eventType == SERVICE_ROUTE_CHANGED {
		cb.onRouteChangedCallback(cb.serviceItem, data)
	}
	CacheService(cb.serviceFinder.config.CachePath, cb.serviceFinder.subscribedService[cb.serviceItem.ServiceName+"_"+cb.serviceItem.ApiVersion])
}

//路由配置信息有改变
func (cb *ServiceChangedCallback) onRouteChangedCallback(service common.ServiceSubscribeItem, data []byte) {
	pushID, routeData, err := common.DecodeValue(data)
	f := &common.ServiceFeedback{
		PushID:       pushID,
		ServiceMete:  cb.serviceFinder.config.MeteData,
		UpdateTime:   time.Now().Unix(),
		UpdateStatus: 1,
		Type:         1,
	}
	if err != nil {
		f.LoadStatus = -1
		go pushServiceFeedback(cb.serviceFinder.config.CompanionUrl, f)
		return
	}
	f.LoadStatus = 1
	f.LoadTime = time.Now().Unix()
	go pushServiceFeedback(cb.serviceFinder.config.CompanionUrl, f)

	serviceId := service.ServiceName + "_" + service.ApiVersion
	serviceRoute := route.ParseRouteData(routeData)
	prevProviderList := cb.serviceFinder.subscribedService[serviceId].ProviderList
	providerMap := cb.serviceFinder.serviceZkData[serviceId].ProviderList
	cb.serviceFinder.serviceZkData[serviceId].Route = serviceRoute
	var maxProviderList []*common.ServiceInstance

	//通过是否有效，先过滤一下服务提供者
	for _, value := range providerMap {
		if value.Config.IsValid {
			maxProviderList = append(maxProviderList, value)
		}
	}

	//根据路由规则来决定最后的服务提供者是那些
	providerList := route.FilterServiceByRouteData(serviceRoute, cb.serviceFinder.config.MeteData.Address, maxProviderList)
	//变更全局数据
	cb.serviceFinder.subscribedService[service.ServiceName+"_"+service.ApiVersion].ProviderList = providerList
	//根据之前的提供者，和目前合法的提供者，来产生相应的事件
	eventList := serviceutil.CompareServiceInstanceList(prevProviderList, providerList)
	if len(eventList) != 0 {
		cb.uh.OnServiceInstanceChanged(service.ServiceName, service.ApiVersion, eventList)
	}

}

func (cb *ServiceChangedCallback) ChildrenChangedCallback(path string, node string, children []string) {
	cb.serviceFinder.locker.Lock()
	defer cb.serviceFinder.locker.Unlock()
	if cb.eventType == SERVICE_INSTANCE_CHANGED {
		cb.OnServiceInstanceChanged(cb.serviceItem, children)
	}
	CacheService(cb.serviceFinder.config.CachePath, cb.serviceFinder.subscribedService[cb.serviceItem.ServiceName+"_"+cb.serviceItem.ApiVersion])
}

//服务的实例的配置发生改变 看is_valid是否被禁用
/**
* 1。无用到可用  ---> 服务提供者可能会变化
* 2。可用到无用  ---> 服务提供者可能会变化
* 3。无用到无用
* 4。可用到可用
 */
func (cb *ServiceChangedCallback) OnServiceInstanceConfigChanged(service common.ServiceSubscribeItem, addr string, data []byte) {
	var serviceId = service.ServiceName + "_" + service.ApiVersion

	pushID, serviceConfData, err := common.DecodeValue(data)
	f := &common.ServiceFeedback{
		PushID:          pushID,
		ServiceMete:     cb.serviceFinder.config.MeteData,
		Provider:        addr,
		ProviderVersion: service.ApiVersion,
		UpdateTime:      time.Now().Unix(),
		UpdateStatus:    1,
		Type:            2,
	}

	if err != nil {
		log.Log.Errorf("parse value err %s", err)
		f.LoadStatus = -1
		go pushServiceFeedback(cb.serviceFinder.config.CompanionUrl, f)
		return
	}
	f.LoadStatus = 1
	f.LoadTime = time.Now().Unix()
	go pushServiceFeedback(cb.serviceFinder.config.CompanionUrl, f)
	serviceConf := serviceutil.ParseServiceConfigData(serviceConfData)
	prevConfig := cb.serviceFinder.serviceZkData[serviceId].ProviderList[addr].Config

	if prevConfig.IsValid == serviceConf.IsValid && strings.Compare(prevConfig.UserConfig, serviceConf.UserConfig) == 0 {
		log.Log.Infof("service instance list not change")
		return
	}
	cb.serviceFinder.serviceZkData[serviceId].ProviderList[addr].Config = serviceConf

	providerList := cb.serviceFinder.subscribedService[serviceId].ProviderList
	var isPrevProvider bool = false
	//处理从可用到无用的变化
	for index, provider := range providerList {
		if strings.Compare(provider.Addr, addr) == 0 {
			isPrevProvider = true
			if !serviceConf.IsValid {
				//之前在服务提供者中，现在不在了。。 服务从可用变为不可用了
				cb.serviceFinder.subscribedService[serviceId].ProviderList = append(providerList[:index], providerList[index+1:]...) //调用
				eventProvider := provider.Dumplication()
				eventProvider.Config.UserConfig = serviceConf.UserConfig
				evetn := common.ServiceInstanceChangedEvent{EventType: common.INSTANCEREMOVE, ServerList: []*common.ServiceInstance{eventProvider}}
				cb.uh.OnServiceInstanceChanged(service.ServiceName, service.ApiVersion, []*common.ServiceInstanceChangedEvent{&evetn})
			}
		}
	}

	//处理从无用到可用的变化
	var shouldAdd bool = true

	if serviceConf.IsValid && !isPrevProvider {
		//之前不在提供者中，现在根据route信息来决定是否放入服务提供者中
		serviceRoutes := cb.serviceFinder.serviceZkData[serviceId].Route.RouteItem
		for _, route := range serviceRoutes {
			providers := route.Provider
			for _, provider := range providers {
				if strings.Compare(provider, addr) == 0 && strings.Compare(route.Only, "Y") == 0 {
					shouldAdd = false
					//在路由组中，且该路由组的only为 YES。。所以跳过该通知
					log.Log.Infof("in route ,not add")
				}
			}
		}
		if shouldAdd {
			serviceInstance := common.ServiceInstance{Addr: addr, Config: serviceConf}
			//增加服务提供者
			cb.serviceFinder.subscribedService[serviceId].ProviderList = append(providerList, &serviceInstance)
			evetn := common.ServiceInstanceChangedEvent{EventType: common.INSTANCEADDED, ServerList: []*common.ServiceInstance{serviceInstance.Dumplication()}}
			cb.uh.OnServiceInstanceChanged(service.ServiceName, service.ApiVersion, []*common.ServiceInstanceChangedEvent{&evetn})
		}

	}

	if strings.Compare(prevConfig.UserConfig, serviceConf.UserConfig) != 0 {
		cb.uh.OnServiceInstanceConfigChanged(service.ServiceName, service.ApiVersion, addr, &common.ServiceInstanceConfig{IsValid: serviceConf.IsValid, UserConfig: serviceConf.UserConfig})
	}

}

//服务的全局配置发生变化，很简单，直接透传就行
func (cb *ServiceChangedCallback) OnServiceConfigChanged(service common.ServiceSubscribeItem, data []byte) {
	pushID, configData, err := common.DecodeValue(data)
	f := &common.ServiceFeedback{
		PushID:       pushID,
		ServiceMete:  cb.serviceFinder.config.MeteData,
		UpdateTime:   time.Now().Unix(),
		UpdateStatus: 1,
		Type:         0,
	}
	//
	if err != nil {
		log.Log.Errorf("pushID：%v unmarsh data err %v", pushID, err)
		f.LoadStatus = -1
		go pushServiceFeedback(cb.serviceFinder.config.CompanionUrl, f)

		return
	}
	f.LoadStatus = 1
	f.LoadTime = time.Now().Unix()
	go pushServiceFeedback(cb.serviceFinder.config.CompanionUrl, f)
	prevConfig := cb.serviceFinder.subscribedService[service.ServiceName+"_"+service.ApiVersion].Config.JsonConfig
	if strings.Compare(prevConfig, string(configData)) == 0 {
		log.Log.Infof("service instance config data not change")

		return
	}
	cb.serviceFinder.subscribedService[service.ServiceName+"_"+service.ApiVersion].Config = &common.ServiceConfig{JsonConfig: string(configData)}
	cb.serviceFinder.serviceZkData[service.ServiceName+"_"+service.ApiVersion].Config = &common.ServiceConfig{JsonConfig: string(configData)}
	cb.uh.OnServiceConfigChanged(service.ServiceName, service.ApiVersion, &common.ServiceConfig{JsonConfig: string(configData)})

}
func (cb *ServiceChangedCallback) Process(path string, node string) {

}
func (cb *ServiceChangedCallback) ChildDeleteCallBack(path string) {
	cb.serviceFinder.locker.Lock()
	defer cb.serviceFinder.locker.Unlock()
	providerPath := strings.Split(path, "/")
	provider := providerPath[len(providerPath)-1]
	var eventList []*common.ServiceInstanceChangedEvent
	var serviceInstance = common.ServiceInstance{Addr: provider, Config: &common.ServiceInstanceConfig{IsValid: false, UserConfig: ""}}
	var event = common.ServiceInstanceChangedEvent{common.INSTANCEREMOVE, []*common.ServiceInstance{&serviceInstance}}
	eventList = append(eventList, &event)
	cb.uh.OnServiceInstanceChanged(cb.serviceItem.ServiceName, cb.serviceItem.ApiVersion, eventList)
}
func getAddProviderAddrList(prevProviderMap map[string]*common.ServiceInstance, currentProviderList []string) []string {
	var addProviderAddrList = make([]string, 0)
	for _, providerAddr := range currentProviderList {
		if _, ok := prevProviderMap[providerAddr]; !ok {
			addProviderAddrList = append(addProviderAddrList, providerAddr)
		}
	}
	return addProviderAddrList
}
func getRemoveProviderAddrList(prevProviderMap map[string]*common.ServiceInstance, currentProviderList []string) []string {
	var removeProviderAddrList = make([]string, 0)
	tempMap := make(map[string]string)
	for _, addr := range currentProviderList {
		tempMap[addr] = addr
	}
	for providerAddr, _ := range prevProviderMap {
		if _, ok := tempMap[providerAddr]; !ok {
			//在之前的提供者中，不在现在的
			removeProviderAddrList = append(removeProviderAddrList, providerAddr)
		}
	}
	return removeProviderAddrList
}

//实例的数量有增加或者减少 没有推送ID 。。则不进行反馈
func (cb *ServiceChangedCallback) OnServiceInstanceChanged(serviceItem common.ServiceSubscribeItem, addrList []string) {

	serviceId := serviceItem.ServiceName + "_" + serviceItem.ApiVersion
	providerMap := cb.serviceFinder.serviceZkData[serviceId].ProviderList
	log.Log.Debugf("current provider list：%v", addrList)
	log.Log.Debugf("current cache provider list：%v", cb.serviceFinder.serviceZkData[serviceId].ProviderList)

	// 当一个节点的回话失效的时候，其所对应的全部节点都会失效。一下子会有多个节点改变
	event := make([]*common.ServiceInstanceChangedEvent, 0)

	//获取多的提供者实例
	addProviderList := getAddProviderAddrList(providerMap, addrList)
	log.Log.Debugf("new provider list is ：%v", addProviderList)
	if len(addProviderList) != 0 {
		//有新增的服务提供者
		rootPath := cb.serviceFinder.rootPath + "/" + serviceItem.ServiceName + "/" + serviceItem.ApiVersion + "/provider"
		callback := NewServiceChangedCallback(serviceItem, SERVICE_INSTANCE_CONFIG_CHANGED, cb.serviceFinder, cb.uh)

		serviceInstanceList := cb.serviceFinder.getServiceInstanceByAddrList(addProviderList, rootPath, &callback)
		var filterInstanceList = make([]*common.ServiceInstance, 0)
		for _, instance := range serviceInstanceList {
			providerMap[instance.Addr] = instance
			if instance.Config != nil && !instance.Config.IsValid {
				continue
			}
			filterInstanceList = append(filterInstanceList, instance)
		}
		resultList := route.FilterServiceByRouteData(cb.serviceFinder.serviceZkData[serviceId].Route, cb.serviceFinder.config.MeteData.Address, filterInstanceList)
		if len(resultList) == 0 {
			log.Log.Infof("new provider ,but route filter")

		} else {
			cb.serviceFinder.subscribedService[serviceId].ProviderList = append(cb.serviceFinder.subscribedService[serviceId].ProviderList, resultList...)
			addEvent := getAddInstanceEvent(resultList)
			event = append(event, addEvent)
		}
	}
	//看是否有服务提供者减小
	removeProviderList := getRemoveProviderAddrList(providerMap, addrList)
	log.Log.Debugf("delete provider list：%v", removeProviderList)

	changeProviderList := make([]*common.ServiceInstance, 0)
	if len(removeProviderList) != 0 {

		for _, addr := range removeProviderList {
			//从原有的提供者中删除
			delete(providerMap, addr)
			visibleProviderList := cb.serviceFinder.subscribedService[serviceId].ProviderList
			for index, provider := range visibleProviderList {
				if strings.Compare(provider.Addr, addr) == 0 {
					cb.serviceFinder.subscribedService[serviceId].ProviderList = append(visibleProviderList[:index], visibleProviderList[index+1:]...)
					changeProviderList = append(changeProviderList, provider)
					break
				}
			}

		}
		if len(changeProviderList) != 0 {
			removeEvent := getRemoveInstnceEvent(changeProviderList)
			event = append(event, removeEvent)
		}
	}
	if len(event) != 0 {
		//通知
		log.Log.Debugf("event notify %v", event)
		cb.uh.OnServiceInstanceChanged(serviceItem.ServiceName, serviceItem.ApiVersion, event)
	}
	CacheService(cb.serviceFinder.config.CachePath, cb.serviceFinder.subscribedService[serviceId])

}
func getRemoveInstnceEvent(insts []*common.ServiceInstance) *common.ServiceInstanceChangedEvent {
	event := common.ServiceInstanceChangedEvent{EventType: common.INSTANCEREMOVE, ServerList: make([]*common.ServiceInstance, 0)}
	for _, inst := range insts {
		event.ServerList = append(event.ServerList, inst.Dumplication())
	}
	return &event
}
func getAddInstanceEvent(insts []*common.ServiceInstance) *common.ServiceInstanceChangedEvent {
	event := common.ServiceInstanceChangedEvent{EventType: common.INSTANCEADDED, ServerList: make([]*common.ServiceInstance, 0)}
	for _, inst := range insts {
		event.ServerList = append(event.ServerList, inst.Dumplication())
	}
	return &event
}

type ConfigChangedCallback struct {
	name         string
	eventType    string
	grayGroupId  string
	uh           common.ConfigChangedHandler
	bootCfg      *common.BootConfig
	sm           storage.StorageManager
	root         string
	configFinder *ConfigFinder
}

func NewConfigChangedCallback(serviceName string, watchType string, rootPath string, userHandle common.ConfigChangedHandler, bootConfig *common.BootConfig, storageMgr storage.StorageManager, configFinder *ConfigFinder) ConfigChangedCallback {
	return ConfigChangedCallback{
		name:         serviceName,
		eventType:    watchType,
		root:         rootPath,
		uh:           userHandle,
		bootCfg:      bootConfig,
		sm:           storageMgr,
		configFinder: configFinder,
	}
}

func (cb *ConfigChangedCallback) ChildDeleteCallBack(path string) {

}
func (cb *ConfigChangedCallback) Process(path string, node string) {
	if strings.HasSuffix(path, "/gray") {
		//如果是gray节点数据改变
		data, err := cb.sm.GetDataWithWatchV2(path, cb)
		if err != nil {
			log.Log.Infof(" [ Process] 从 %s  %s", path, " 获取数据失败")
			return
		}
		cb.OnGrayConfigChanged(cb.name, data)
		return
	}

	var currentGrayGroupId string
	if groupId, ok := cb.configFinder.grayConfig.Load(cb.configFinder.config.MeteData.Address); ok {
		currentGrayGroupId = groupId.(string)
	}

	if len(currentGrayGroupId) == 0 && strings.Contains(path, "/gray/") {
		log.Log.Infof("当前不在灰度组，但是通知是属于灰度组的，不进行处理")
		return
	}
	if len(currentGrayGroupId) != 0 && !strings.Contains(path, "/"+currentGrayGroupId) {
		log.Log.Infof("当前在灰度组，但是通知是属于其他灰度组的，不进行处理")
		return
	}
	var isSubscribeFile bool
	for _, value := range cb.configFinder.fileSubscribe {
		if strings.Compare(cb.name, value) == 0 {
			isSubscribeFile = true
		}
	}
	if !isSubscribeFile {
		log.Log.Infof("不是订阅的文件，不进行推送")
		return
	}

	data, err := cb.sm.GetDataWithWatchV2(path, cb)
	if err != nil {
		log.Log.Infof(" [ Process] 从 %s,%s", path, " 获取数据失败")
		return
	}
	cb.OnConfigFileChanged(cb.name, data, path)
}
func (cb *ConfigChangedCallback) DataChangedCallback(path string, node string, data []byte) {

	if cb.eventType == CONFIG_CHANGED {
		cb.OnConfigFileChanged(cb.name, data, path)
	}

}

func (cb *ConfigChangedCallback) ChildrenChangedCallback(path string, node string, children []string) {

}

func (cb *ConfigChangedCallback) OnGrayConfigChanged(name string, data []byte) {
	var currentGrayGroupId string
	var prevGrayGroupId string
	consumerPath := cb.configFinder.rootPath + "/consumer"
	if grayConfig, ok := ParseGrayConfigData(cb.bootCfg.MeteData.Address, data); ok {

		if groupId, ok := cb.configFinder.grayConfig.Load(cb.configFinder.config.MeteData.Address); ok {
			prevGrayGroupId = groupId.(string)
		}
		if groupId, ok := grayConfig[cb.configFinder.config.MeteData.Address]; ok {
			currentGrayGroupId = groupId
		}
		cb.configFinder.grayConfig.Store(cb.configFinder.config.MeteData.Address, currentGrayGroupId)
		if strings.Compare(prevGrayGroupId, currentGrayGroupId) == 0 {
			//如果之前的group和现在的一样，则代表没有切换灰度组。直接结束
			return
		} else if len(currentGrayGroupId) != 0 {
			//当前在灰度组
			//不相等，则代表灰度组有改变。需要重新获取节点配置信息
			if len(prevGrayGroupId) == 0 {
				removePath := consumerPath + "/normal/" + cb.configFinder.config.MeteData.Address
				cb.sm.Remove(removePath)
			} else {
				removePath := consumerPath + "/gray/" + prevGrayGroupId + "/" + cb.configFinder.config.MeteData.Address
				cb.sm.Remove(removePath)
			}
			consumerPath += "/gray/" + currentGrayGroupId + "/" + cb.configFinder.config.MeteData.Address
			cb.sm.SetTempPath(consumerPath)
			f := cb.configFinder
			for _, fileName := range f.fileSubscribe {
				callback := NewConfigChangedCallback(fileName, CONFIG_CHANGED, f.rootPath, cb.uh, f.config, f.storageMgr, f)
				basePath := cb.root + "/gray/" + currentGrayGroupId + "/" + fileName
				data, err := cb.sm.GetDataWithWatchV2(basePath, &callback)
				if err != nil {
					if err.Error() == common.ZK_NODE_DOSE_NOT_EXIST {
						log.Log.Infof(" [OnGrayConfigChanged] 重新从路径 %s ,%s ,%s", basePath, " 获取配置失败 ", err)
						var errInfo = common.ConfigErrInfo{FileName: fileName, ErrCode: 0, ErrMsg: "配置文件不存在"}
						cb.uh.OnError(errInfo)
						return
					}
					log.Log.Infof(" [OnGrayConfigChanged] 重新从路径 %s ,%s ,%s", basePath, " 获取配置失败 ", err)
					return
				}
				cb.OnConfigFileChanged(fileName, data, basePath)
			}

		} else {
			removePath := consumerPath + "/gray/" + prevGrayGroupId + "/" + cb.configFinder.config.MeteData.Address
			cb.sm.Remove(removePath)
			consumerPath += "/normal/" + cb.configFinder.config.MeteData.Address
			cb.sm.SetTempPath(consumerPath)
			f := cb.configFinder
			for _, fileName := range f.fileSubscribe {
				callback := NewConfigChangedCallback(fileName, CONFIG_CHANGED, f.rootPath, cb.uh, f.config, f.storageMgr, f)
				basePath := cb.root + "/" + fileName
				data, err := cb.sm.GetDataWithWatchV2(basePath, &callback)
				if err != nil {
					if err.Error() == common.ZK_NODE_DOSE_NOT_EXIST {
						var errInfo = common.ConfigErrInfo{FileName: fileName, ErrCode: 0, ErrMsg: "配置文件不存在"}
						cb.uh.OnError(errInfo)
						return
					}
					log.Log.Infof(" [OnGrayConfigChanged] 重新从路径 %s ,%s ,%s", basePath, " 获取配置失败 ", err)
					return
				}
				cb.OnConfigFileChanged(fileName, data, basePath)
			}

		}

	}

}
func (cb *ConfigChangedCallback) OnConfigFileChanged(name string, data []byte, path string) {

	var currentGrayGroupId string
	if groupId, ok := cb.configFinder.grayConfig.Load(cb.configFinder.config.MeteData.Address); ok {
		currentGrayGroupId = groupId.(string)
	} else {
		currentGrayGroupId = "0"
	}
	pushID, file, err := common.DecodeValue(data)
	if err != nil {
		f := &common.ConfigFeedback{
			PushID:       pushID,
			ServiceMete:  cb.bootCfg.MeteData,
			Config:       name,
			UpdateTime:   time.Now().Unix(),
			UpdateStatus: 1,
			GrayGroupId:  currentGrayGroupId,
			LoadStatus:   -1,
			LoadTime:     time.Now().Unix(),
		}

		go pushConfigFeedback(cb.bootCfg.CompanionUrl, f)
	} else {
		f := &common.ConfigFeedback{
			PushID:       pushID,
			ServiceMete:  cb.bootCfg.MeteData,
			Config:       name,
			UpdateTime:   time.Now().Unix(),
			UpdateStatus: 1,
			LoadStatus:   1,
			GrayGroupId:  currentGrayGroupId,
			LoadTime:     time.Now().Unix(),
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

		res := cb.uh.OnConfigFileChanged(c)
		if res == false {
			f.LoadStatus = -1
		}
		go pushConfigFeedback(cb.bootCfg.CompanionUrl, f)
		go CacheConfig(cb.bootCfg.CachePath, c)

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

func pushService(companionUrl string, project string, group string, service string, apiVersion string) error {
	url := companionUrl + "/finder/register_service_info"
	return companion.RegisterService(hc, url, project, group, service, apiVersion)
}
