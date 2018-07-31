package finder

import (
	"encoding/json"
	"fmt"
	"sync"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/storage"
	storagecommon "git.xfyun.cn/AIaaS/finder-go/storage/common"
	"git.xfyun.cn/AIaaS/finder-go/utils/stringutil"
)

type ServiceFinder struct {
	locker     sync.Mutex
	rootPath   string
	config     *common.BootConfig
	storageMgr storage.StorageManager
	//当前服务下有哪些提供者
	usedService       map[string]*common.Service
	subscribedService map[string]*common.Service
	mutex             sync.Mutex
}

func NewServiceFinder(root string, bc *common.BootConfig, sm storage.StorageManager) *ServiceFinder {
	finder := &ServiceFinder{
		locker:            sync.Mutex{},
		rootPath:          root,
		config:            bc,
		storageMgr:        sm,
		usedService:       make(map[string]*common.Service, 0),
		subscribedService: make(map[string]*common.Service, 0),
	}

	return finder
}

func (f *ServiceFinder) RegisterService() error {
	return f.registerService(f.config.MeteData.Address)
}

func (f *ServiceFinder) RegisterServiceWithAddr(addr string) error {
	return f.registerService(addr)
}

func (f *ServiceFinder) UnRegisterService() error {
	servicePath := fmt.Sprintf("%s/%s/provider/%s", f.rootPath, f.config.MeteData.Service, f.config.MeteData.Address)

	return f.storageMgr.RemoveInRecursive(servicePath)
}

func (f *ServiceFinder) UnRegisterServiceWithAddr(addr string) error {
	servicePath := fmt.Sprintf("%s/%s/provider/%s", f.rootPath, f.config.MeteData.Service, addr)

	return f.storageMgr.RemoveInRecursive(servicePath)
}

func (f *ServiceFinder) UseService(name []string) (map[string]*common.Service, error) {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  errors.ServiceMissName,
			Func: "UseService",
		}

		return nil, err
	}

	f.locker.Lock()
	defer f.locker.Unlock()

	var addrList []string
	serviceList := make(map[string]*common.Service)
	for _, n := range name {
		if s, ok := f.usedService[n]; ok {
			serviceList[n] = s
			continue
		}

		servicePath := fmt.Sprintf("%s/%s/provider", f.rootPath, n)
		logger.Info("useservice:", servicePath)
		//获取服务提供者
		addrList, err = f.storageMgr.GetChildren(servicePath)
		if err != nil {
			logger.Info("useservice:", err)
			service, err := GetServiceFromCache(f.config.CachePath, n)
			if err != nil {
				logger.Error(err)
				//todo notify
				return nil, err
			}

			serviceList[n] = service
			f.usedService[n] = service
		} else if len(addrList) > 0 {
			logger.Info("servicePath:", servicePath)
			logger.Info(addrList)
			serviceList[n] = f.getService(servicePath, n, addrList)
			err = CacheService(f.config.CachePath, serviceList[n])
			if err != nil {
				logger.Error("CacheService failed")
			}
		}

		f.usedService[n] = serviceList[n]

		err = f.registerConsumer(n, f.config.MeteData.Address)
		if err != nil {
			logger.Error("registerConsumer failed,", err)
		}
	}

	return serviceList, err
}

func (f *ServiceFinder) UseAndSubscribeService(name []string, handler common.ServiceChangedHandler) (map[string]*common.Service, error) {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  errors.ServiceMissName,
			Func: "UseAndSubscribeService",
		}

		return nil, err
	}

	f.locker.Lock()
	defer f.locker.Unlock()

	serviceList := make(map[string]*common.Service)
	for _, n := range name {
		if s, ok := f.subscribedService[n]; ok {
			serviceList[n] = s
			continue
		}

		servicePath := fmt.Sprintf("%s/%s/provider", f.rootPath, n)
		callback := NewServiceChangedCallback(n, SERVICE_INSTANCE_CHANGED, f.rootPath, handler, f.config, f.storageMgr)
		addrList, err := f.storageMgr.GetChildrenWithWatch(servicePath, &callback)
		if err != nil {
			logger.Error("f.storageMgr.GetChildrenWithWatch:", err)
			service, err := GetServiceFromCache(f.config.CachePath, n)
			if err != nil {
				logger.Info("GetServiceFromCache ", err)
				return nil, err
			}

			serviceList[n] = service
			f.subscribedService[n] = service
		}

		service, err := f.getServiceWithWatcher(servicePath, n, addrList, handler)
		if err != nil {
			return nil, err
		}

		if len(service.ServerList) > 0 {
			err = CacheService(f.config.CachePath, service)
			if err != nil {
				logger.Info("CacheService failed")
			}
		} else {
			service, err = GetServiceFromCache(f.config.CachePath, n)
			if err != nil {
				logger.Info(err)
				return nil, err
				//todo notify
			}
		}

		serviceList[n] = service
		f.subscribedService[n] = service

		err = f.registerConsumer(n, f.config.MeteData.Address)
		if err != nil {
			logger.Error("registerConsumer failed,", err)
		}

		// zkutil.ServiceEventPool.Append(common.ServiceEventPrefix+n, interHandle)
	}

	return serviceList, nil
}

func (f *ServiceFinder) UnSubscribeService(name string) error {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  errors.ServiceMissName,
			Func: "UnSubscribeService",
		}
		return err
	}

	f.locker.Lock()
	defer f.locker.Unlock()

	delete(f.subscribedService, name)

	return nil
}

func (f *ServiceFinder) registerService(addr string) error {
	if stringutil.IsNullOrEmpty(addr) {
		err := &errors.FinderError{
			Ret:  errors.ServiceMissAddr,
			Func: "RegisterService",
		}

		logger.Error("RegisterService:", err)
		return err
	}

	data, err := getDefaultServiceItemConfig(addr)
	if err != nil {
		logger.Error("RegisterService->getDefaultServiceItemConfig:", err)
		return err
	}
	parentPath := fmt.Sprintf("%s/%s/provider", f.rootPath, f.config.MeteData.Service)
	err = f.register(parentPath, addr, data)
	if err != nil {
		logger.Error("RegisterService->register:", err)
		return err
	}

	err = pushService(f.config.CompanionUrl, f.config.MeteData.Project, f.config.MeteData.Group, f.config.MeteData.Service)
	if err != nil {
		logger.Error("RegisterService->registerService:", err)
	}

	return nil
}

func (f *ServiceFinder) registerConsumer(service string, addr string) error {
	if stringutil.IsNullOrEmpty(addr) {
		err := &errors.FinderError{
			Ret:  errors.ServiceMissAddr,
			Func: "registerConsumer",
		}

		logger.Error("registerConsumer:", err)
		return err
	}

	data, err := getDefaultConsumerItemConfig(addr)
	if err != nil {
		logger.Error("registerConsumer->getDefaultConsumerItemConfig:", err)
		return err
	}
	parentPath := fmt.Sprintf("%s/%s/consumer", f.rootPath, service)
	err = f.register(parentPath, addr, data)
	if err != nil {
		logger.Error("registerConsumer->register:", err)
		return err
	}

	return nil
}

func (f *ServiceFinder) register(parentPath string, addr string, data []byte) error {
	logger.Info("call register func")
	servicePath := parentPath + "/" + addr
	logger.Info("servicePath:", servicePath)

	return f.storageMgr.SetTempPath(servicePath)
}

func getDefaultServiceItemConfig(addr string) ([]byte, error) {
	defaultServiceInstanceConfig := common.ServiceInstanceConfig{
		Weight:  100,
		IsValid: true,
	}

	data, err := json.Marshal(defaultServiceInstanceConfig)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var encodedData []byte
	encodedData, err = common.EncodeValue("", data)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return encodedData, nil
}

func getDefaultConsumerItemConfig(addr string) ([]byte, error) {
	defaultConsumeInstanceConfig := common.ConsumerInstanceConfig{
		IsValid: true,
	}

	data, err := json.Marshal(defaultConsumeInstanceConfig)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var encodedData []byte
	encodedData, err = common.EncodeValue("", data)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return encodedData, nil
}

func getServiceInstance(sm storage.StorageManager, path string, addr string) (*common.ServiceInstance, error) {
	data, err := sm.GetData(path + "/" + addr)
	if err != nil {
		return nil, err
	}

	var item []byte
	_, item, err = common.DecodeValue(data)
	if err != nil {
		return nil, err
	}

	logger.Info(string(item))
	serviceInstanceConfig := &common.ServiceInstanceConfig{}
	err = json.Unmarshal(item, serviceInstanceConfig)
	if err != nil {
		return nil, err
	}

	serviceInstance := new(common.ServiceInstance)
	serviceInstance.Addr = addr
	serviceInstance.Config = serviceInstanceConfig

	return serviceInstance, nil
}

func (f *ServiceFinder) getService(servicePath string, name string, addrList []string) *common.Service {
	var service = &common.Service{Name: name, ServerList: make([]*common.ServiceInstance, 0), Config: &common.ServiceConfig{}}
	for _, addr := range addrList {
		serviceInstance, err := getServiceInstance(f.storageMgr, servicePath, addr)
		if err != nil {
			logger.Info(err)
			// todo
			continue
		}

		service.ServerList = append(service.ServerList, serviceInstance)
	}
	// todo
	service.Config.ProxyMode = "default"
	service.Config.LoadBalanceMode = "default"

	return service
}

func (f *ServiceFinder) getServiceWithWatcher(servicePath string, name string, addrList []string, handler common.ServiceChangedHandler) (*common.Service, error) {
	var service = &common.Service{Name: name, ServerList: make([]*common.ServiceInstance, 0), Config: &common.ServiceConfig{}}
	for _, addr := range addrList {
		callback := NewServiceChangedCallback(name, SERVICE_INSTANCE_CONFIG_CHANGED, f.rootPath, handler, f.config, f.storageMgr)
		serviceInstance, err := f.getServiceInstanceWithWatcher(servicePath, addr, &callback)
		if err != nil {
			logger.Info(err)
			// todo
			return nil, err
		}

		service.ServerList = append(service.ServerList, serviceInstance)
	}
	// todo
	service.Config.ProxyMode = "default"
	service.Config.LoadBalanceMode = "default"

	return service, nil
}

func (f *ServiceFinder) getServiceInstanceWithWatcher(servicePath string, addr string, callback storagecommon.ChangedCallback) (*common.ServiceInstance, error) {
	data, err := f.storageMgr.GetDataWithWatch(servicePath+"/"+addr, callback)
	if err != nil {
		return nil, err
	}

	_, item, err := common.DecodeValue(data)
	if err != nil {
		return nil, err
	}

	serviceInstance := &common.ServiceInstance{Addr: addr, Config: new(common.ServiceInstanceConfig)}
	err = json.Unmarshal(item, serviceInstance.Config)
	if err != nil {
		return nil, err
	}

	return serviceInstance, nil
}
