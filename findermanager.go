package finder

import (
	"finder-go/common"
	"finder-go/errors"
	"finder-go/utils/fileutil"
	"finder-go/utils/stringutil"
	"finder-go/utils/zkutil"
	"os"
)

// FinderManager for controll all
type FinderManager struct {
	config        *common.BootConfig
	ConfigFinder  *ConfigFinder
	ServiceFinder *ServiceFinder
	zkManager     *zkutil.ZkManager
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

// NewFinder for creating an instance
func NewFinder(config common.BootConfig) (*FinderManager, error) {
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
	fm.zkManager, err = zkutil.NewZkManager(fm.config)
	if err != nil {
		return nil, err
	}
	fm.ConfigFinder = &ConfigFinder{zkManager: fm.zkManager}
	fm.ServiceFinder = &ServiceFinder{zkManager: fm.zkManager, config: fm.config}

	if err != nil {
		return nil, err
	}

	return fm, nil
}

func onCfgUpdateEvent(c common.Config) int {
	return errors.ConfigSuccess
}
