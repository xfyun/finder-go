package finder

import (
	"strings"
	"sync"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/storage"
	"git.xfyun.cn/AIaaS/finder-go/utils/fileutil"
	"log"
)

var (
	configEventPrefix = "config_"
)

type ConfigFinder struct {
	locker           sync.Mutex
	rootPath         string
	currentWatchPath string
	config           *common.BootConfig
	storageMgr       storage.StorageManager
	usedConfig       sync.Map
	fileSubscribe    []string
	grayConfig       sync.Map
}

func NewConfigFinder(root string, bc *common.BootConfig, sm storage.StorageManager) *ConfigFinder {

	finder := &ConfigFinder{
		locker:     sync.Mutex{},
		rootPath:   root,
		config:     bc,
		storageMgr: sm,
		usedConfig: sync.Map{},
	}

	return finder
}

// UseConfig for 订阅相关配置文件
func (f *ConfigFinder) UseConfig(name []string) (map[string]*common.Config, error) {
	if len(name) == 0 {
		err := errors.NewFinderError(errors.ConfigMissName)
		return nil, err
	}

	f.locker.Lock()
	defer f.locker.Unlock()
	err := GetGrayConfigData(f, f.rootPath, nil)
	if err != nil {
		logger.Info("获取灰度配置信息出错", err)
		return nil, err
	}
	configFiles := make(map[string]*common.Config)
	for _, n := range name {
		if c, ok := f.usedConfig.Load(name); !ok {
			//先获取gray的数据，用于判断订阅的配置是否在灰度组中
			basePath := f.rootPath
			if groupId, ok := f.grayConfig.Load(f.config.MeteData.Address); ok {
				basePath += "/gray/" + groupId.(string)
			}
			//真正获取数据
			data, err := f.storageMgr.GetData(basePath + "/" + n)
			if err != nil {
				//出错 从配置文件中取
				onUseConfigErrorWithCache(configFiles, n, f.config.CachePath, err)
			} else {
				_, fData, err := common.DecodeValue(data)
				if err != nil {
					//出错 从配置文件中获取
					onUseConfigErrorWithCache(configFiles, n, f.config.CachePath, err)
				} else {
					var config *common.Config
					if fileutil.IsTomlFile(n) {
						tomlConfig := fileutil.ParseTomlFile(fData)
						config = &common.Config{Name: n, File: fData, ConfigMap: tomlConfig}
					} else {
						config = &common.Config{Name: n, File: fData}
					}
					configFiles[n] = config
					//存到缓存
					err = CacheConfig(f.config.CachePath, config)
					if err != nil {
						logger.Error("CacheConfig:", err)
					}
				}
			}
		} else {
			// todo
			if config, ok := c.(common.Config); ok {
				configFiles[n] = &config
			} else {
				// get config from cache
				configFiles[n] = getCachedConfig(n, f.config.CachePath)
			}
		}
	}

	return configFiles, nil
}

// UseAndSubscribeConfig for
//新增监控灰度组的Watch
func (f *ConfigFinder) UseAndSubscribeConfig(name []string, handler common.ConfigChangedHandler) (map[string]*common.Config, error) {
	if len(name) == 0 {
		err := errors.NewFinderError(errors.ConfigMissName)
		return nil, err
	}
	f.locker.Lock()
	defer f.locker.Unlock()


	//先查看灰度组的设置
	callback := NewConfigChangedCallback(f.config.MeteData.Address, CONFIG_CHANGED, f.rootPath, handler, f.config, f.storageMgr, f)
	err := GetGrayConfigData(f, f.rootPath, &callback)
	if err != nil {
		logger.Info("获取灰度配置信息出错", err)
		return nil, err
	}


	if groupId, ok := f.grayConfig.Load(f.config.MeteData.Address); ok {
		//如果在灰度组。则进行注册到灰度组中
		if ok:=f.checkFileExist(f.rootPath+"/gray/"+ groupId.(string),name);!ok {
			log.Println("订阅的文件中，有不存在的，不进行订阅,path: ",f.rootPath+"/gray/"+ groupId.(string))
			return nil,errors.NewFinderError(errors.ConfigFileNotExist)
		}
	} else {
		if ok:=f.checkFileExist(f.rootPath,name);!ok {
			log.Println("订阅的文件中，有不存在的，不进行订阅,path : ",f.rootPath)
			return nil,errors.NewFinderError(errors.ConfigFileNotExist)
		}
	}

	consumerPath := f.rootPath + "/consumer"
	if groupId, ok := f.grayConfig.Load(f.config.MeteData.Address); ok {
		//如果在灰度组。则进行注册到灰度组中
		consumerPath += "/gray/" + groupId.(string) + "/" + f.config.MeteData.Address
		f.storageMgr.SetTempPath(consumerPath)
	} else {
		consumerPath += "/normal/" + f.config.MeteData.Address
		f.storageMgr.SetTempPath(consumerPath)
	}

	configFiles := make(map[string]*common.Config)
	path := ""
	for _, n := range name {
		f.fileSubscribe = append(f.fileSubscribe, n)
		if c, ok := f.usedConfig.Load(n); ok {
			// todo
			if config, ok := c.(common.Config); ok {
				configFiles[n] = &config
			} else {
				// get config from cache
				configFiles[n] = getCachedConfig(n, f.config.CachePath)
			}

			continue
		} else {
			basePath := f.rootPath
			if groupId, ok := f.grayConfig.Load(f.config.MeteData.Address); ok {
				basePath += "/gray/" + groupId.(string)
			}
			callback := NewConfigChangedCallback(n, CONFIG_CHANGED, f.rootPath, handler, f.config, f.storageMgr, f)

			//根据获取的灰度组设置的结果，来到特定的节点获取配置文件数据
			path = basePath + "/" + n
			data, err := f.storageMgr.GetDataWithWatchV2(path, &callback)
			if err != nil {
				if strings.Compare(err.Error(),common.ZK_NODE_DOSE_NOT_EXIST)==0{
					log.Println("配置文件不存在，请先配置文件。文件:",name)
					return nil,errors.NewFinderError(errors.ConfigFileNotExist)
				}
				onUseConfigErrorWithCache(configFiles, n, f.config.CachePath, err)

			} else {
				_, fData, err := common.DecodeValue(data)
				if err != nil {
					onUseConfigErrorWithCache(configFiles, n, f.config.CachePath, err)
				} else {
					//
					confMap := make(map[string]interface{})
					if fileutil.IsTomlFile(n) {
						confMap = fileutil.ParseTomlFile(fData)
					}
					config := &common.Config{Name: n, File: fData, ConfigMap: confMap}
					configFiles[n] = config
					f.usedConfig.Store(n, config)
					//放到文件中
					err = CacheConfig(f.config.CachePath, config)
					if err != nil {
						logger.Error("CacheConfig:", err)
					}
				}
			}

		}
	}
	log.Println("订阅的文件是：",f.fileSubscribe)
	return configFiles, nil
}

func (f* ConfigFinder) checkFileExist(basePath string,names []string) bool{
	//TODO 判断文件是否存在，不存在则直接报错，

	files,err:=f.storageMgr.GetChildren(basePath)
	if err != nil {
		log.Println("获取配置文件出错，直接返回",err)
		return false
	}
	if len(names) >len(files) {
		log.Println("当前有的配置文件为：",files," 要订阅的配置文件为：",names,",两者不匹配！")
		return false
	}
	for _,subFileName :=range names{
		var isExist =false
		for _,existFile :=range files{
			if existFile==subFileName{
				isExist=true
			}
		}
		if !isExist {
			log.Println("当前有的配置文件为：",files," 要订阅的配置文件 ",subFileName," 不存在！")
			return false
		}
	}
	return true

}
func (f *ConfigFinder) UnSubscribeConfig(name string) error {
	var err error
	if len(name) == 0 {
		err = errors.NewFinderError(errors.ConfigMissName)
		return err
	}
	for index, value := range f.fileSubscribe {
		if strings.Compare(name, value) == 0 {
			f.fileSubscribe = append(f.fileSubscribe[:index], f.fileSubscribe[index+1:]...)
		}
	}
	if len(f.fileSubscribe) == 0 {
		f.removeConfigConsumer()
	}

	return nil
}

func (f *ConfigFinder)removeConfigConsumer(){
	//如果订阅文件的个数为0，则取消注册者
	consumerPath := f.rootPath + "/consumer"
	if groupId, ok := f.grayConfig.Load(f.config.MeteData.Address); ok {
		//如果在灰度组。则进行注册到灰度组中
		consumerPath += "/gray/" + groupId.(string) + "/" + f.config.MeteData.Address
		f.storageMgr.Remove(consumerPath)
	} else {
		consumerPath += "/normal/" + f.config.MeteData.Address
		f.storageMgr.Remove(consumerPath)
	}
}
func (f *ConfigFinder) BatchUnSubscribeConfig (names []string)error{
	if len(names)==0 {
		err := errors.NewFinderError(errors.ConfigMissName)
		return err
	}
	for _,name :=range names{
		for index, value := range f.fileSubscribe {
			if strings.Compare(name, value) == 0 {
				f.fileSubscribe = append(f.fileSubscribe[:index], f.fileSubscribe[index+1:]...)
			}
		}
	}
	if len(f.fileSubscribe) == 0 {
		f.removeConfigConsumer()
	}
	return nil
}
// onUseConfigError with cache
func onUseConfigErrorWithCache(configFiles map[string]*common.Config, name string, cachePath string, err error) {
	logger.Error("onUseConfigError:", err)
	configFiles[name] = getCachedConfig(name, cachePath)
}

func getCachedConfig(name string, cachePath string) *common.Config {
	config, err := GetConfigFromCache(cachePath, name)
	if err != nil {
		logger.Error("GetConfigFromCache:", err)
		return nil
	}

	return config
}
