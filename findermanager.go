package finder

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	companion "git.xfyun.cn/AIaaS/finder-go/companion"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	log "git.xfyun.cn/AIaaS/finder-go/log"
	"git.xfyun.cn/AIaaS/finder-go/storage"
	"git.xfyun.cn/AIaaS/finder-go/utils/arrayutil"
	"git.xfyun.cn/AIaaS/finder-go/utils/fileutil"
	"git.xfyun.cn/AIaaS/finder-go/utils/netutil"
	"git.xfyun.cn/AIaaS/finder-go/utils/stringutil"
	"sync"
)

var (
	hc *http.Client
)

const VERSION = "2.0.4"

type zkAddrChangeCallback struct {
	path string
	fm   *FinderManager
}

func (callback *zkAddrChangeCallback) ChildDeleteCallBack(path string) {

}
func (callback *zkAddrChangeCallback) ChildrenChangedCallback(path string, node string, children []string) {

}
func (callback *zkAddrChangeCallback) Process(path string, node string) {

	log.Log.Debug("zk_node_path节点事件处理", path)
	fm := callback.fm
	storageMgr, storageCfg, err := initStorageMgr(fm.config)
	if err != nil {
		log.Log.Error("zk信息出错，重新尝试", err)
		go watchStorageInfo(fm)
	} else {
		go watchZkInfo(fm)
		fm.storageMgr = storageMgr
		fm.ConfigFinder.storageMgr = storageMgr
		fm.ConfigFinder.rootPath = storageCfg.ConfigRootPath
		fm.ConfigFinder.config = fm.config
		fm.ServiceFinder.storageMgr = storageMgr
		fm.ServiceFinder.rootPath = storageCfg.ServiceRootPath
		if len(fm.ServiceFinder.subscribedService) != 0 {
			ReGetServiceInfo(fm)
		}
		if len(fm.ConfigFinder.fileSubscribe) != 0 {
			ReGetConfigInfo(fm)
		}
	}
}

func (callback *zkAddrChangeCallback) DataChangedCallback(path string, node string, data []byte) {

}

func init() {
	hc = &http.Client{
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

}

// FinderManager for controll all
type FinderManager struct {
	config        *common.BootConfig
	storageMgr    storage.StorageManager
	ConfigFinder  *ConfigFinder
	ServiceFinder *ServiceFinder
}

func checkCachePath(path string) (string, error) {
	if stringutil.IsNullOrEmpty(path) {
		p, err := os.Getwd()
		if err == nil {
			p += (fileutil.GetSystemSeparator() + common.DefaultCacheDir)
			path = p
		} else {
			return path, err
		}
	}

	return path, nil
}

func createCacheDir(path string) error {
	exist, err := fileutil.ExistPath(path)
	if err == nil && !exist {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

func checkConfig(c *common.BootConfig) {
	if c.ExpireTimeout <= 0 {
		c.ExpireTimeout = 3 * time.Second
	}
}

func getStorageInfo(config *common.BootConfig) (*common.StorageInfo, error) {
	url := config.CompanionUrl + fmt.Sprintf("/finder/query_zk_info?project=%s&group=%s&service=%s&version=%s", config.MeteData.Project, config.MeteData.Group, config.MeteData.Service, config.MeteData.Version)
	info, err := companion.GetStorageInfo(hc, url)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func checkAddr(n []string, o []string) bool {
	vchanged := false
	for _, nv := range o {
		if !arrayutil.Contains(nv, o) {
			vchanged = true
		}
	}

	return vchanged
}

func onZkInfoChanged(smr storage.StorageManager) {
	// todo.
}

func getStorageConfig(config *common.BootConfig) (*storage.StorageConfig, error) {
	checkConfig(config)
	info, err := getStorageInfo(config)
	if err != nil {
		return nil, err
	}
	storageConfig := &storage.StorageConfig{
		Name:   "zookeeper",
		Params: make(map[string]string),
	}

	storageConfig.Params["servers"] = strings.Join(info.Addr, ",")
	storageConfig.Params["session_timeout"] = strconv.FormatInt(int64(config.ExpireTimeout/time.Millisecond), 10)
	storageConfig.Params["zk_node_path"] = info.ZkNodePath
	storageConfig.ConfigRootPath = info.ConfigRootPath
	storageConfig.ServiceRootPath = info.ServiceRootPath

	return storageConfig, nil
}

func initStorageMgr(config *common.BootConfig) (storage.StorageManager, *storage.StorageConfig, error) {
	storageConfig, err := getStorageConfig(config)
	if err != nil {
		log.Log.Error("[ initStorageMgr ] getStorageConfig:", err)
		return nil, nil, err
	}
	log.Log.Debug("storageConfig信息：", storageConfig.Params)
	storageMgr, err := storage.NewManager(storageConfig)
	if err != nil {
		log.Log.Error("[ initStorageMgr ] NewManager:", err)
		return nil, storageConfig, err
	}
	err = storageMgr.Init()
	if err != nil {
		log.Log.Error("[ initStorageMgr ] Init err", err)
		return nil, storageConfig, err
	}

	return storageMgr, storageConfig, nil
}

// NewFinder for creating an instance
func newFinder(config common.BootConfig) (*FinderManager, error) {
	log.Log = log.NewDefaultLogger()
	if stringutil.IsNullOrEmpty(config.CompanionUrl) {
		err := errors.NewFinderError(errors.MissCompanionUrl)
		return nil, err
	}

	if stringutil.IsNullOrEmpty(config.MeteData.Address) {
		localIP, err := netutil.GetLocalIP(config.CompanionUrl)
		if err != nil {
			log.Log.Error(err)
			return nil, err
		}
		config.MeteData.Address = localIP
	}

	// 检查缓存路径，如果传入cachePath是空，则使用默认路径
	p, err := checkCachePath(config.CachePath)
	if err != nil {
		return nil, err
	}

	// 创建缓存目录
	err = createCacheDir(p)
	if err != nil {
		return nil, err
	}
	config.CachePath = p
	// 初始化finder
	fm := new(FinderManager)
	fm.config = &config
	// 初始化zk
	var storageCfg *storage.StorageConfig
	fm.storageMgr, storageCfg, err = initStorageMgr(fm.config)
	if err != nil {
		log.Log.Info("初始化zk信息出错，开启新的goroutine 去不断尝试")
		fm.ConfigFinder = NewConfigFinder("", fm.config, nil)
		fm.ServiceFinder = NewServiceFinder("", fm.config, nil)
		//return nil, err
	} else {
		fm.ConfigFinder = NewConfigFinder(storageCfg.ConfigRootPath, fm.config, fm.storageMgr)
		fm.ServiceFinder = NewServiceFinder(storageCfg.ServiceRootPath, fm.config, fm.storageMgr)
	}

	return fm, nil
}

func NewFinderWithLogger(config common.BootConfig, logger log.Logger) (*FinderManager, error) {
	if logger == nil {
		log.Log = log.NewDefaultLogger()
	} else {
		log.Log = logger
	}
	log.Log.Info("current version : " + VERSION)
	if stringutil.IsNullOrEmpty(config.CompanionUrl) {
		err := errors.NewFinderError(errors.MissCompanionUrl)
		return nil, err
	}

	if stringutil.IsNullOrEmpty(config.MeteData.Address) {
		localIP, err := netutil.GetLocalIP(config.CompanionUrl)
		if err != nil {
			log.Log.Error(err)
			return nil, err
		}
		config.MeteData.Address = localIP
	}

	// 检查缓存路径，如果传入cachePath是空，则使用默认路径
	p, err := checkCachePath(config.CachePath)
	if err != nil {
		return nil, err
	}
	// 创建缓存目录
	err = createCacheDir(p)
	if err != nil {
		return nil, err
	}
	config.CachePath = p
	// 初始化finder
	fm := new(FinderManager)
	fm.config = &config
	// 初始化zk
	var storageCfg *storage.StorageConfig
	fm.storageMgr, storageCfg, err = initStorageMgr(fm.config)
	if err != nil {
		log.Log.Info("初始化zk信息出错，开启新的goroutine 去不断尝试")
		fm.ConfigFinder = NewConfigFinder("", fm.config, nil)
		fm.ServiceFinder = NewServiceFinder("", fm.config, nil)
		go watchStorageInfo(fm)
		return fm, nil
	} else {
		fm.ConfigFinder = NewConfigFinder(storageCfg.ConfigRootPath, fm.config, fm.storageMgr)
		fm.ServiceFinder = NewServiceFinder(storageCfg.ServiceRootPath, fm.config, fm.storageMgr)
	}
	//创建一个goroutine来执行监听zk地址的数据
	go watchZkInfo(fm)
	return fm, nil
}

func watchZkInfo(fm *FinderManager) {

	zkNodePath, err := fm.storageMgr.GetZkNodePath()
	if err != nil {
		log.Log.Error("zk的节点信息为空")
	}
	log.Log.Debug("zk的节点信息为:", zkNodePath)
	fm.storageMgr.GetDataWithWatchV2(zkNodePath, &zkAddrChangeCallback{path: zkNodePath, fm: fm})
}

func watchStorageInfo(fm *FinderManager) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	var tempPathMap sync.Map
	if fm.storageMgr != nil {
		tempPathMap = fm.storageMgr.GetTempPaths()
	}
	for {
		select {
		case <-ticker.C:

			getStorageConfig(fm.config)
			storageMgr, storageCfg, err := initStorageMgr(fm.config)
			if err != nil {
				fm.storageMgr = nil
				fm.ConfigFinder.storageMgr = nil
				fm.ServiceFinder.storageMgr = nil
				log.Log.Info("初始化zk信息出错，重新尝试  ", err)
			} else {
				fm.storageMgr = storageMgr
				fm.ConfigFinder.storageMgr = storageMgr
				fm.ConfigFinder.rootPath = storageCfg.ConfigRootPath
				fm.ConfigFinder.config = fm.config

				fm.ServiceFinder.storageMgr = storageMgr
				fm.ServiceFinder.rootPath = storageCfg.ServiceRootPath
			}
		}
		if fm.storageMgr != nil {

			go watchZkInfo(fm)
			fm.storageMgr.SetTempPaths(tempPathMap)
			fm.storageMgr.RecoverTempPaths()

			if len(fm.ServiceFinder.subscribedService) != 0 {
				log.Log.Debug("重新拉取订阅的服务信息，", fm.ServiceFinder.subscribedService)
				//重新拉取所有订阅服务的信息
				ReGetServiceInfo(fm)
			}
			if len(fm.ConfigFinder.fileSubscribe) != 0 {
				log.Log.Debug("重新拉取订阅的配置文件信息，", fm.ConfigFinder.fileSubscribe)
				ReGetConfigInfo(fm)
			}

			break
		}
	}
}
func ReGetConfigInfo(fm *FinderManager) {
	handler := fm.ConfigFinder.handler
	var fileSubscribe []string
	for _,file :=range fm.ConfigFinder.fileSubscribe{
		fileSubscribe=append(fileSubscribe,file)
	}
	fm.ConfigFinder.fileSubscribe=[]string{}
	fm.ConfigFinder.grayConfig.Range( func(key, value interface{}) bool{
		fm.ConfigFinder.grayConfig.Delete(key)
		return true
	})


	fm.ConfigFinder.usedConfig.Range( func(key, value interface{}) bool{
		fm.ConfigFinder.usedConfig.Delete(key)
		return true
	})
	fileMap, err := fm.ConfigFinder.UseAndSubscribeConfig(fileSubscribe, handler)
	if err != nil {
		log.Log.Error("获取信息失败", err)
	}
	for _, fileData := range fileMap {
		var config = common.Config{Name: fileData.Name, File: fileData.File, ConfigMap: fileData.ConfigMap}
		if handler != nil {
			handler.OnConfigFileChanged(&config)

		}
	}
}

func ReGetServiceInfo(fm *FinderManager) {
	for _, value := range fm.ServiceFinder.subscribedService {
		var item = common.ServiceSubscribeItem{ServiceName: value.ServiceName, ApiVersion: value.ApiVersion}
		servicePath := fmt.Sprintf("%s/%s/%s", fm.ServiceFinder.rootPath, item.ServiceName, item.ApiVersion)
		service, err := fm.ServiceFinder.getServiceWithWatcher(servicePath, item, fm.ServiceFinder.handler)
		if err != nil {
			log.Log.Error("获取信息失败", err)
		}
		if service == nil {
			log.Log.Debug("获取不到service")
		}
		cacheService, err := GetServiceFromCache(fm.config.CachePath, item)
		ChangeEvent(cacheService, service, fm.ServiceFinder.handler)
		if service != nil {
			CacheService(fm.config.CachePath, service)
		}

	}
}

func ChangeEvent(prevService *common.Service, currService *common.Service, handler common.ServiceChangedHandler) {
	if prevService == nil {
		handler.OnServiceConfigChanged(currService.ServiceName, currService.ApiVersion, &common.ServiceConfig{JsonConfig: currService.Config.JsonConfig})
		eventList := providerChangeEvent([]*common.ServiceInstance{}, currService.ProviderList)
		if len(eventList) == 0 {
			return
		}
		handler.OnServiceInstanceChanged(currService.ServiceName, currService.ApiVersion, eventList)
		return
	}
	if currService == nil {
		handler.OnServiceConfigChanged(prevService.ServiceName, prevService.ApiVersion, &common.ServiceConfig{JsonConfig: prevService.Config.JsonConfig})
		eventList := providerChangeEvent(prevService.ProviderList, []*common.ServiceInstance{})
		if len(eventList) == 0 {
			return
		}
		handler.OnServiceInstanceChanged(currService.ServiceName, currService.ApiVersion, eventList)
		return
	}
	prevConfig := prevService.Config
	currConfig := currService.Config
	if prevConfig.JsonConfig != currConfig.JsonConfig {
		handler.OnServiceConfigChanged(currService.ServiceName, currService.ApiVersion, &common.ServiceConfig{JsonConfig: currConfig.JsonConfig})
	}
	eventList := providerChangeEvent(prevService.ProviderList, currService.ProviderList)
	if len(eventList) == 0 {
		return
	}
	handler.OnServiceInstanceChanged(currService.ServiceName, currService.ApiVersion, eventList)
	return
}

func providerChangeEvent(prevProviderList, currProviderList []*common.ServiceInstance) []*common.ServiceInstanceChangedEvent {
	var eventList []*common.ServiceInstanceChangedEvent
	if len(prevProviderList) == 0 && len(currProviderList) == 0 {
		return nil
	}
	if len(prevProviderList) == 0 {
		var changeList []*common.ServiceInstance
		for _, provider := range currProviderList {
			changeList = append(changeList, provider.Dumplication())
		}
		event := common.ServiceInstanceChangedEvent{EventType: common.INSTANCEADDED, ServerList: changeList}
		eventList = append(eventList, &event)
		return eventList
	}
	if len(currProviderList) == 0 {
		var changeList []*common.ServiceInstance
		for _, provider := range prevProviderList {
			changeList = append(changeList, provider.Dumplication())
		}
		event := common.ServiceInstanceChangedEvent{EventType: common.INSTANCEREMOVE, ServerList: changeList}
		eventList = append(eventList, &event)
		return eventList
	}
	var addServerList []*common.ServiceInstance
	//TODO 后续优化
	var providerMap = make(map[string]*common.ServiceInstance)
	for _, prevProvider := range prevProviderList {
		providerMap[prevProvider.Addr] = prevProvider
	}
	for _, currProvider := range currProviderList {
		if _, ok := providerMap[currProvider.Addr]; !ok {
			addServerList = append(addServerList, currProvider.Dumplication())
		} else {
			delete(providerMap, currProvider.Addr)
		}
	}
	var removeServerList []*common.ServiceInstance
	for _, provider := range providerMap {
		removeServerList = append(removeServerList, provider.Dumplication())
	}
	removeEvent := common.ServiceInstanceChangedEvent{EventType: common.INSTANCEREMOVE, ServerList: removeServerList}
	eventList = append(eventList, &removeEvent)
	addEvent := common.ServiceInstanceChangedEvent{EventType: common.INSTANCEADDED, ServerList: addServerList}
	eventList = append(eventList, &addEvent)
	return eventList

}

func DestroyFinder(finder *FinderManager) {
	finder.storageMgr.Destroy()
	// todo
}

func onCfgUpdateEvent(c common.Config) int {
	return errors.ConfigSuccess
}
