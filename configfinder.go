package finder

import (
	"sync"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/storage"
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
		err := &errors.FinderError{
			Ret:  errors.ConfigMissName,
			Func: "UseConfig",
		}

		return nil, err
	}

	f.locker.Lock()
	defer f.locker.Unlock()

	configFiles := make(map[string]*common.Config)
	for _, n := range name {
		if c, ok := f.usedConfig.Load(name); !ok {
			data, err := f.storageMgr.GetData(f.rootPath + "/" + n)
			if err != nil {
				onUseConfigErrorWithCache(configFiles, n, f.config.CachePath, err)
			} else {
				_, fData, err := common.DecodeValue(data)
				if err != nil {
					onUseConfigErrorWithCache(configFiles, n, f.config.CachePath, err)
				} else {
					config := &common.Config{Name: n, File: fData}
					configFiles[n] = config

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
func (f *ConfigFinder) UseAndSubscribeConfig(name []string, handler common.ConfigChangedHandler) (map[string]*common.Config, error) {
	if len(name) == 0 {
		err := &errors.FinderError{
			Ret:  errors.ConfigMissName,
			Func: "UseConfig",
		}

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
			path = f.rootPath + "/" + n
			callback := NewConfigChangedCallback(n, CONFIG_CHANGED, f.rootPath, handler, f.config, f.storageMgr)
			data, err := f.storageMgr.GetDataWithWatch(path, &callback)

			if err != nil {
				onUseConfigErrorWithCache(configFiles, n, f.config.CachePath, err)
			} else {
				//对配置信息进行解码
				_, fData, err := common.DecodeValue(data)
				if err != nil {
					onUseConfigErrorWithCache(configFiles, n, f.config.CachePath, err)
				} else {
					config := &common.Config{Name: n, File: fData}
					configFiles[n] = config
					//缓存到内存
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
		err = &errors.FinderError{
			Ret:  errors.ConfigMissName,
			Func: "UnSubscribeConfig",
		}
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
