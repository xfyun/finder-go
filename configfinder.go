package finder

import (
	"sync"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/storage"
	"git.xfyun.cn/AIaaS/finder-go/utils/fileutil"
)

var (
	configEventPrefix = "config_"
)

type ConfigFinder struct {
	locker     sync.Mutex
	rootPath   string
	config     *common.BootConfig
	storageMgr storage.StorageManager
	usedConfig sync.Map
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

// UseConfig for
func (f *ConfigFinder) UseConfig(name []string) (map[string]*common.Config, error) {
	if len(name) == 0 {
		err := errors.NewFinderError(errors.ConfigMissName)
		return nil, err
	}

	f.locker.Lock()
	defer f.locker.Unlock()

	configFiles := make(map[string]*common.Config)
	for _, n := range name {
		if c, ok := f.usedConfig.Load(name); !ok {
			//先获取gray的数据
			basePath := f.rootPath
			if groupId, ok := GetGrayConfigData(f, f.rootPath, nil); ok {
				basePath += "/gray/" + groupId
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

	configFiles := make(map[string]*common.Config)
	path := ""
	for _, n := range name {

		if c, ok := f.usedConfig.Load(name); ok {
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

			//先查看灰度组的设置
			grayCallBack := NewConfigChangedCallback(n, GRAY_CONFIG_CHANGED, f.rootPath, handler, f.config, f.storageMgr)
			if groupId, ok := GetGrayConfigData(f, f.rootPath, &grayCallBack); ok {
				basePath += "/gray/" + groupId
				//TODO 这个地方有点问题
				grayCallBack.grayGroupId = groupId
			}

			//根据获取的灰度组设置的结果，来到特定的节点获取配置文件数据
			path = basePath + "/" + n
			callback := NewConfigChangedCallback(n, CONFIG_CHANGED, f.rootPath, handler, f.config, f.storageMgr)
			data, err := f.storageMgr.GetDataWithWatch(path, &callback)
			logger.Info(path)
			if err != nil {
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

	return configFiles, nil
}

func (f *ConfigFinder) UnSubscribeConfig(name string) error {
	var err error
	if len(name) == 0 {
		err = errors.NewFinderError(errors.ConfigMissName)
		return err
	}

	// todo

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
