package finder

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	companion "git.xfyun.cn/AIaaS/finder-go/companion"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/storage"
	"git.xfyun.cn/AIaaS/finder-go/utils/arrayutil"
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

	logger = common.NewDefaultLogger()
}

// FinderManager for controll all
type FinderManager struct {
	config         *common.BootConfig
	storageMgr     storage.StorageManager
	ConfigFinder   *ConfigFinder
	ServiceFinder  *ServiceFinder
	InternalLogger common.Logger
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

func watchStorageInfo(smr *storage.StorageManager) {
	// for t := range smr.checkZkInfoTicker.C {
	// 	//log.Println(t)
	// 	if t.IsZero() {

	// 	}
	// 	metadata, err := companion.GetStorageInfo(hc, url)
	// 	if err != nil {
	// 		// todo.
	// 		continue
	// 	}
	// 	vchanged := checkAddr(metadata.Addr, zm.MetaData.Addr)
	// 	if vchanged {
	// 		zm.MetaData.ZkAddr = metadata.ZkAddr
	// 		zm.MetaData.ConfigRootPath = metadata.ConfigRootPath
	// 		zm.MetaData.ServiceRootPath = metadata.ServiceRootPath
	// 		// 通知zkinfo更新，执行相关逻辑
	// 		onZkInfoChanged(zm)
	// 	}
	// }
}

func getStorageConfig(config *common.BootConfig) (*storage.StorageConfig, error) {
	checkConfig(config)
	info, err := getStorageInfo(config)
	if err != nil {
		return nil, err
	}
	//zm.checkZkInfoTicker = time.NewTicker(config.TickerDuration)
	// 开启一个协程去检测zkinfo变化
	//go watchStorageInfo(zm)

	storageConfig := &storage.StorageConfig{
		Name:   "zookeeper",
		Params: make(map[string]string),
	}

	storageConfig.Params["servers"] = strings.Join(info.Addr, ",")
	storageConfig.Params["session_timeout"] = strconv.FormatInt(int64(config.ExpireTimeout/time.Millisecond), 10)
	storageConfig.ConfigRootPath = info.ConfigRootPath
	storageConfig.ServiceRootPath = info.ServiceRootPath

	return storageConfig, nil
}

func initStorageMgr(config *common.BootConfig) (storage.StorageManager, *storage.StorageConfig, error) {
	storageConfig, err := getStorageConfig(config)
	if err != nil {
		logger.Error("getStorageConfig:", err)
		return nil, nil, err
	}

	storageMgr, err := storage.NewManager(storageConfig)
	if err != nil {
		return nil, storageConfig, err
	}
	err = storageMgr.Init()
	if err != nil {
		log.Println(err)
		return nil, storageConfig, err
	}

	return storageMgr, storageConfig, nil
}

// NewFinder for creating an instance
func NewFinder(config common.BootConfig) (*FinderManager, error) {
	// logger := common.NewDefaultLogger()
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
	fm.InternalLogger = logger
	fm.config = &config
	// 初始化zk
	var storageCfg *storage.StorageConfig
	fm.storageMgr, storageCfg, err = initStorageMgr(fm.config)
	if err != nil {
		return nil, err
	}

	fm.ConfigFinder = NewConfigFinder(storageCfg.ConfigRootPath, fm.config, fm.storageMgr)
	fm.ServiceFinder = NewServiceFinder(storageCfg.ServiceRootPath, fm.config, fm.storageMgr)

	return fm, nil
}

func NewFinderWithLogger(config common.BootConfig, logger common.Logger) (*FinderManager, error) {
	if logger == nil {
		logger = common.NewDefaultLogger()
	} else {
		logger = logger
	}

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
	fm.InternalLogger = logger
	fm.config = &config
	// 初始化zk
	var storageCfg *storage.StorageConfig
	fm.storageMgr, storageCfg, err = initStorageMgr(fm.config)
	if err != nil {
		return nil, err
	}

	fm.ConfigFinder = NewConfigFinder(storageCfg.ConfigRootPath, fm.config, fm.storageMgr)
	fm.ServiceFinder = NewServiceFinder(storageCfg.ServiceRootPath, fm.config, fm.storageMgr)

	return fm, nil
}

func DestroyFinder(finder *FinderManager) {
	finder.storageMgr.Destroy()
	// todo
}

func onCfgUpdateEvent(c common.Config) int {
	return errors.ConfigSuccess
}
