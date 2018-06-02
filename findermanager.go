package finder

import (
	"net"
	"net/http"
	"os"
	"time"

	"git.xfyun.cn/AIaaS/finder-go/storage"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/utils/fileutil"
	"git.xfyun.cn/AIaaS/finder-go/utils/netutil"
	"git.xfyun.cn/AIaaS/finder-go/utils/stringutil"
)

var (
	hc     *http.Client
	logger common.Logger
)

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
	logger        common.Logger
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

func initCacheDir(path string) (string, error) {
	// 检查缓存路径，如果传入cachePath是空，则使用默认路径
	p, err := checkCachePath(path)
	if err != nil {
		return "", err
	}

	// 创建缓存目录
	err = createCacheDir(p)
	if err != nil {
		return "", err
	}

	return p, nil
}

// NewFinder for creating an instance
func NewFinder(config common.BootConfig) (*FinderManager, error) {
	return NewFinderWithLogger(config, nil)
}

// NewFinderWithLogger for creating an instance with logger
func NewFinderWithLogger(config common.BootConfig, logger common.Logger) (*FinderManager, error) {
	if stringutil.IsNullOrEmpty(config.CompanionUrl) {
		err := &errors.FinderError{
			Ret:  errors.MissCompanionUrl,
			Func: "NewFinder",
		}

		return nil, err
	}

	if stringutil.IsNullOrEmpty(config.MeteData.Address) {
		localIP, err := netutil.GetLocalIP(config.CompanionUrl)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		config.MeteData.Address = localIP
	}

	// 初始化缓存目录
	initCacheDir(config.CachePath)

	// 初始化finder
	fm := new(FinderManager)
	fm.config = &config
	if logger == nil {
		fm.logger = common.NewDefaultLogger()
	} else {
		fm.logger = logger
	}

	// 初始化存储
	storageConfig := &storage.StorageConfig{
		Name:   "",
		Params: nil,
	}
	storageManager, err := storage.NewManager(storageConfig)
	if err != nil {
		return nil, err
	}
	fm.storageMgr = storageManager
	fm.ConfigFinder = NewConfigFinder("", fm.config, fm.storageMgr, fm.logger)
	//fm.ConfigFinder = &ConfigFinder{storageMgr: fm.storageMgr, config: fm.config, logger: fm.logger}
	fm.ServiceFinder = NewServiceFinder("", fm.config, fm.storageMgr, fm.logger)
	//fm.ServiceFinder = &ServiceFinder{storageMgr: fm.storageMgr, config: fm.config, logger: fm.logger, SubscribedService: make(map[string]*common.Service)}

	return fm, nil
}

func DestroyFinder(finder *FinderManager) {
	finder.storageMgr.Destroy()
	// todo
}
