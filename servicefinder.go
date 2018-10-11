package finder

import (
	"encoding/json"
	"fmt"
	"sync"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/log"
	"git.xfyun.cn/AIaaS/finder-go/route"
	"git.xfyun.cn/AIaaS/finder-go/storage"
	"git.xfyun.cn/AIaaS/finder-go/utils/serviceutil"
	"git.xfyun.cn/AIaaS/finder-go/utils/stringutil"
	"strings"
)

type ServiceFinder struct {
	locker            sync.Mutex
	rootPath          string
	config            *common.BootConfig
	handler           common.ServiceChangedHandler
	storageMgr        storage.StorageManager
	usedService       map[string]*common.Service
	subscribedService map[string]*common.Service
	serviceZkData     map[string]*ServiceZkData
	mutex             sync.Mutex
}
type ServiceZkData struct {
	ServiceName string
	ApiVersion  string
	//所有的提供者 key是addr
	ProviderList map[string]*common.ServiceInstance
	Config       *common.ServiceConfig
	Route        *common.ServiceRoute
}

func NewServiceFinder(root string, bc *common.BootConfig, sm storage.StorageManager) *ServiceFinder {
	finder := &ServiceFinder{
		locker:            sync.Mutex{},
		rootPath:          root,
		config:            bc,
		storageMgr:        sm,
		usedService:       make(map[string]*common.Service, 0),
		subscribedService: make(map[string]*common.Service, 0),
		serviceZkData:     make(map[string]*ServiceZkData, 0),
	}

	return finder
}

func (f *ServiceFinder) RegisterServiceWithAddr(addr string, version string) error {
	if f.storageMgr == nil {
		return errors.NewFinderError(errors.ZkConnectionLoss)
	}
	log.Log.Debug("RegisterServiceWithAddr : addr-> ",addr," version->",version)
	return f.registerService(addr, version)
}
func (f *ServiceFinder) RegisterService(version string) error {
	if f.storageMgr == nil {
		return errors.NewFinderError(errors.ZkConnectionLoss)
	}
	log.Log.Debug("RegisterServiceWithAddr : version->",version)
	return f.registerService(f.config.MeteData.Address, version)
}
func (f *ServiceFinder) UnRegisterService(version string) error {
	if f.storageMgr == nil {
		return errors.NewFinderError(errors.ZkConnectionLoss)
	}
	servicePath := fmt.Sprintf("%s/%s/%s/provider/%s", f.rootPath, f.config.MeteData.Service, version, f.config.MeteData.Address)
	return f.storageMgr.RemoveInRecursive(servicePath)
}

func (f *ServiceFinder) UnRegisterServiceWithAddr(version string, addr string) error {
	if f.storageMgr == nil {
		return errors.NewFinderError(errors.ZkConnectionLoss)
	}
	servicePath := fmt.Sprintf("%s/%s/%s/provider/%s", f.rootPath, f.config.MeteData.Service, version, addr)

	return f.storageMgr.RemoveInRecursive(servicePath)
}

func (f *ServiceFinder) UseService(serviceItems []common.ServiceSubscribeItem) (map[string]*common.Service, error) {
	var err error
	if len(serviceItems) == 0 {
		err = errors.NewFinderError(errors.ServiceMissItem)
		return nil, err
	}

	f.locker.Lock()
	defer f.locker.Unlock()

	serviceList := make(map[string]*common.Service)
	for _, item := range serviceItems {
		//这个usedService 是作何用处？
		serviceId := item.ServiceName + "_" + item.ApiVersion
		if service, ok := f.usedService[serviceId]; ok {
			serviceList[serviceId] = service
			continue
		}
		//测试用
		servicePath := fmt.Sprintf("/polaris/service/05127d76c3a6fe7c3375562921560a20/%s/%s", item.ServiceName, item.ApiVersion)
		//servicePath := fmt.Sprintf("%s/%s/%s", f.rootPath, item.ServiceName, item.ApiVersion)
		log.Log.Info(" useservice:", servicePath)
		serviceList[serviceId], err = f.getService(servicePath, item)
		//存入缓存文件
		err = CacheService(f.config.CachePath, serviceList[serviceId])
		if err != nil {
			log.Log.Error("CacheService failed")
		}

		//	f.usedService[n] = serviceList[n]

		// err = f.registerConsumer(n, f.config.MeteData.Address)
		// if err != nil {
		// 	logger.Error("registerConsumer failed,", err)
		// }
	}

	return serviceList, err
}

func (f *ServiceFinder) UseAndSubscribeService(serviceItems []common.ServiceSubscribeItem, handler common.ServiceChangedHandler) (map[string]common.Service, error) {
	var err error
	if len(serviceItems) == 0 {
		err = errors.NewFinderError(errors.ServiceMissItem)
		return nil, err
	}

	f.locker.Lock()
	defer f.locker.Unlock()
	f.handler = handler
	serviceList := make(map[string]common.Service)

	if f.storageMgr == nil {
		if !f.config.CacheService {
			log.Log.Info(" [ UseAndSubscribeService ] 不使用缓存，直接退出")
			return nil, nil
		}
		log.Log.Info(" [ UseAndSubscribeService ] 从缓存中获取数据")

		//说明zk信息目前有误，暂时使用缓存数据
		for _, item := range serviceItems {
			serviceId := item.ServiceName + "_" + item.ApiVersion
			service, err := GetServiceFromCache(f.config.CachePath, item)
			if err != nil {
				log.Log.Info("从缓存中获取该服务失败，服务为：", serviceId)
				f.subscribedService[serviceId] = &common.Service{ServiceName: item.ServiceName, ApiVersion: item.ApiVersion}

			} else {
				serviceList[serviceId] = service.Dumplication()
				f.subscribedService[serviceId] = service
			}

		}
		return serviceList, nil
	}
	for _, item := range serviceItems {
		serviceId := item.ServiceName + "_" + item.ApiVersion
		if service, ok := f.subscribedService[serviceId]; ok {
			serviceList[serviceId] = service.Dumplication()
			continue
		}
		servicePath := fmt.Sprintf("%s/%s/%s", f.rootPath, item.ServiceName, item.ApiVersion)
		service, err := f.getServiceWithWatcher(servicePath, item, handler)
		if err != nil {
			log.Log.Info(" [ UseAndSubscribeService ] 订阅【服务名:", item.ServiceName, ",版本号:", item.ApiVersion, "】 出错", err)
			continue
		}
		if service == nil {
			continue
		}
		serviceList[serviceId] = service.Dumplication()
		f.subscribedService[serviceId] = service

		err = f.registerConsumer(item, f.config.MeteData.Address)
		if err != nil {
			log.Log.Error("registerConsumer failed,", err)
		}
		CacheService(f.config.CachePath, f.subscribedService[serviceId])
	}
	return serviceList, nil
}

func (f *ServiceFinder) UnSubscribeService(name string) error {
	var err error
	if len(name) == 0 {
		//	err = errors.NewFinderError(errors.ServiceMissName)
		return err
	}

	f.locker.Lock()
	defer f.locker.Unlock()

	delete(f.subscribedService, name)

	return nil
}

func (f *ServiceFinder) registerService(addr string, apiVersion string) error {
	if stringutil.IsNullOrEmpty(addr) {
		err := errors.NewFinderError(errors.ServiceMissAddr)
		return err
	}
	if stringutil.IsNullOrEmpty(apiVersion) {
		log.Log.Info("[registerService] 缺失apiVersion数据")
		return errors.NewFinderError(errors.ServiceMissApiVersion)
	}
	//目前不考虑目录不存在的情况
	path := fmt.Sprintf("%s/%s/%s/provider/%s", f.rootPath, f.config.MeteData.Service, apiVersion, addr)
	log.Log.Debug("registerService -> path -> ",path)
	err := f.storageMgr.SetTempPath(path)
	if err != nil {
		log.Log.Info("服务注册失败", err)
		return err
	}
	go pushService(f.config.CompanionUrl, f.config.MeteData.Project, f.config.MeteData.Group, f.config.MeteData.Service, apiVersion)
	//if err != nil {
	//	logger.Error("RegisterService->registerService:", err)
	//}
	return nil
}

func (f *ServiceFinder) registerConsumer(service common.ServiceSubscribeItem, addr string) error {
	if stringutil.IsNullOrEmpty(addr) {
		err := errors.NewFinderError(errors.ServiceMissAddr)
		log.Log.Error("registerConsumer:", err)
		return err
	}

	parentPath := fmt.Sprintf("%s/%s/%s/consumer", f.rootPath, service.ServiceName, service.ApiVersion)
	err := f.register(parentPath, addr)
	if err != nil {
		log.Log.Error("registerConsumer->register:", err)
		return err
	}

	return nil
}
func (f *ServiceFinder) getServiceInstanceByAddrList(providerAddrList []string, rootPath string, handler *ServiceChangedCallback) []*common.ServiceInstance {
	var serviceInstanceList = make([]*common.ServiceInstance, 0)
	for _, providerAddr := range providerAddrList {
		log.Log.Debug(" [ getServiceInstanceByAddrList] providerAddr:", providerAddr, " rootPath :", rootPath)
		service, err := getServiceInstance(f.storageMgr, rootPath, providerAddr, handler)
		if err != nil || service == nil {
			continue
		}
		serviceInstanceList = append(serviceInstanceList, service)
	}
	return serviceInstanceList
}
func (f *ServiceFinder) register(parentPath string, addr string) error {
	log.Log.Info("call register func")
	servicePath := parentPath + "/" + addr
	log.Log.Info("servicePath:", servicePath)
	return f.storageMgr.SetTempPath(servicePath)
}

func getDefaultServiceItemConfig(addr string) ([]byte, error) {
	defaultServiceInstanceConfig := common.ServiceInstanceConfig{
		IsValid: true,
	}

	data, err := json.Marshal(defaultServiceInstanceConfig)
	if err != nil {
		log.Log.Error(err)
		return nil, err
	}

	var encodedData []byte
	encodedData, err = common.EncodeValue("", data)
	if err != nil {
		log.Log.Error(err)
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
		log.Log.Error(err)
		return nil, err
	}

	var encodedData []byte
	encodedData, err = common.EncodeValue("", data)
	if err != nil {
		log.Log.Error(err)
		return nil, err
	}

	return encodedData, nil
}

func getServiceInstance(sm storage.StorageManager, path string, addr string, callback *ServiceChangedCallback) (*common.ServiceInstance, error) {
	var data []byte
	var err error
	if callback != nil {
		data, err = sm.GetDataWithWatch(path+"/"+addr, callback)
	} else {
		data, err = sm.GetData(path + "/" + addr)
	}
	if err != nil {
		log.Log.Info("从 ", path+"/"+addr, " 获取数据出错 ", err)
		//TODO 是否需要返回默认的
		return nil, err
	}
	serviceInstance := new(common.ServiceInstance)
	//解析数据
	if data == nil || len(data) == 0 {
		//获取数据为空
		log.Log.Info("从 ", path+"/"+addr, " 获取数据为空")
		serviceInstance.Config = getDefaultServiceInstanceConfig()
	} else {
		//获取的提供者配置数据不为空
		var item []byte
		_, item, err = common.DecodeValue(data)
		if err != nil {
			log.Log.Info("实例上的配置数据不符合规范，反序列化出错", err)
			//使用默认的配置
			serviceInstance.Config = getDefaultServiceInstanceConfig()
		} else {
			serviceInstance.Config = serviceutil.ParseServiceConfigData(item)
		}

	}
	serviceInstance.Addr = addr
	log.Log.Info("-------------", serviceInstance.Config)
	return serviceInstance, nil
}
func getDefaultServiceInstanceConfig() *common.ServiceInstanceConfig {
	serviceInstanceConfig := &common.ServiceInstanceConfig{}
	serviceInstanceConfig.IsValid = true
	serviceInstanceConfig.UserConfig = ""
	return serviceInstanceConfig
}

func (f *ServiceFinder) getService(servicePath string, serviceItem common.ServiceSubscribeItem) (*common.Service, error) {
	var service = &common.Service{ServiceName: serviceItem.ServiceName, ApiVersion: serviceItem.ApiVersion, ProviderList: make([]*common.ServiceInstance, 0), Config: &common.ServiceConfig{}}
	var serviceZkData = &ServiceZkData{ServiceName: serviceItem.ServiceName, ApiVersion: serviceItem.ApiVersion, ProviderList: make(map[string]*common.ServiceInstance)}
	var providerPath = servicePath + "/provider"
	var confPath = servicePath + "/conf"
	var routePath = servicePath + "/route"
	//先找provider路径下的数据
	providerList, err := f.storageMgr.GetChildren(providerPath)
	if err != nil {
		log.Log.Info("从path: ", providerPath, " 获取服务提供者出错", err)
		return nil, nil
	}
	for _, providerAddr := range providerList {
		serviceInstance, err := getServiceInstance(f.storageMgr, providerPath, providerAddr, nil)
		if err != nil {
			//TODO 当data为nil的时候，会返回错误。。这里要处理一下
			log.Log.Info("获取提供者实例信息出错，path=:", providerPath+"/"+providerAddr, " 错误为:", err)
			// todo
			continue
		}
		//如果该提供者被禁用了，则跳过
		if !serviceInstance.Config.IsValid {
			continue
		}
		service.ProviderList = append(service.ProviderList, serviceInstance)
	}
	//获取config下的信息
	confData, err := f.storageMgr.GetData(confPath)
	if err != nil {
		log.Log.Info("从path: ", confPath, " 获取配置数据出错", err)
	} else {
		_, fData, err := common.DecodeValue(confData)
		if err != nil {
			log.Log.Info("解析配置数据出错", err)
		}
		service.Config = &common.ServiceConfig{JsonConfig: string(fData)}
	}

	//获取route数据
	routeData, err := f.storageMgr.GetData(routePath)
	if err != nil {
		log.Log.Info("从path: ", routePath, " 获取路由数据出错", err)
	} else if routeData != nil {
		_, fData, err := common.DecodeValue(routeData)
		if err != nil {
			log.Log.Info("解析路由数据出错", err)
		}
		log.Log.Info(`{"RouteItem":` + string(fData) + "}")

		var serviceRoute common.ServiceRoute
		json.Unmarshal([]byte(`{"RouteItem":`+string(fData)+"}"), &serviceRoute)
		log.Log.Info(serviceRoute)
		serviceZkData.Route = &serviceRoute

		//使用route进行过滤数据
		service.ProviderList = route.FilterServiceByRouteData(serviceZkData.Route, f.config.MeteData.Address, service.ProviderList)
	}

	log.Log.Info(service)
	return service, nil
}

func (f *ServiceFinder) getServiceWithWatcher(servicePath string, serviceItem common.ServiceSubscribeItem, handler common.ServiceChangedHandler) (*common.Service, error) {
	var service = &common.Service{ServiceName: serviceItem.ServiceName, ApiVersion: serviceItem.ApiVersion, ProviderList: make([]*common.ServiceInstance, 0)}

	var serviceZkData = &ServiceZkData{ServiceName: serviceItem.ServiceName, ApiVersion: serviceItem.ApiVersion, ProviderList: make(map[string]*common.ServiceInstance)}
	f.serviceZkData[serviceItem.ServiceName+"_"+serviceItem.ApiVersion] = serviceZkData
	var providerPath = servicePath + "/provider"
	var confPath = servicePath + "/conf"
	var routePath = servicePath + "/route"
	//先找provider路径下的数据
	callback := NewServiceChangedCallback(serviceItem, SERVICE_INSTANCE_CHANGED, f, handler)
	//获取数据的时候添加子节点变更的Watcher
	providerList, err := f.storageMgr.GetChildrenWithWatch(providerPath, &callback)

	//TODO 提供者为空的情况
	if err != nil {
		if strings.Compare("zk: node does not exist", err.Error()) == 0 {
			//节点不存在，则新建之
			err := f.storageMgr.SetPath(providerPath)
			if err != nil {
				log.Log.Info("[ GetChildrenWithWatch ] 创建节点: ", providerPath)
			}
			return nil, err
		}
		log.Log.Info("从path: ", providerPath, " 获取服务提供者出错", err)
		return nil, err
	}
	if len(providerList) == 0 {
		log.Log.Info(" [ getServiceWithWatcher ]目前没有服务提供者存在")
	}
	for _, providerAddr := range providerList {
		proiderCallBack := NewServiceChangedCallback(serviceItem, SERVICE_INSTANCE_CONFIG_CHANGED, f, handler)
		serviceInstance, err := getServiceInstance(f.storageMgr, providerPath, providerAddr, &proiderCallBack)
		if err != nil {
			//TODO 当data为nil的时候，会返回错误。。这里要处理一下
			log.Log.Info("获取提供者实例信息出错，path=:", providerPath+"/"+providerAddr, " 错误为:", err)
			// todo
			continue
		}
		serviceZkData.ProviderList[serviceInstance.Addr] = serviceInstance
		//如果该提供者被禁用了，则跳过
		if serviceInstance.Config != nil && !serviceInstance.Config.IsValid {
			continue
		}
		service.ProviderList = append(service.ProviderList, serviceInstance)
	}

	//获取config下的信息
	confCallBack := NewServiceChangedCallback(serviceItem, SERVICE_CONFIG_CHANGED, f, handler)
	confData, err := f.storageMgr.GetDataWithWatch(confPath, &confCallBack)
	if err != nil {

		log.Log.Info("从path: ", confPath, " 获取配置数据出错", err)
		if strings.Compare(common.ZK_NODE_DOSE_NOT_EXIST, err.Error()) == 0 {
			log.Log.Info("新建节点: ", confPath)
			f.storageMgr.SetPath(confPath)
		}
		service.Config = &common.ServiceConfig{JsonConfig: ""}

	} else if len(confData) == 0 {
		service.Config = &common.ServiceConfig{JsonConfig: ""}
		log.Log.Info("从path: ", confPath, " 获取配置为空，没有对应的配置信息")
	} else {
		_, fData, err := common.DecodeValue(confData)
		if err != nil {
			log.Log.Info("解析配置数据出错", err)
		}
		service.Config = &common.ServiceConfig{JsonConfig: string(fData)}
		serviceZkData.Config = &common.ServiceConfig{JsonConfig: string(fData)}
	}
	//获取route数据
	routeCallBack := NewServiceChangedCallback(serviceItem, SERVICE_ROUTE_CHANGED, f, handler)
	routeData, err := f.storageMgr.GetDataWithWatch(routePath, &routeCallBack)
	if err != nil {
		log.Log.Info("从path: ", routePath, " 获取路由数据出错", err)
		if strings.Compare(common.ZK_NODE_DOSE_NOT_EXIST, err.Error()) == 0 {
			log.Log.Info("新建节点: ", routePath)
			f.storageMgr.SetPath(routePath)
		}
		serviceZkData.Route = &common.ServiceRoute{RouteItem: []*common.RouteItem{}}

	} else if routeData != nil && len(routeData) == 0 {
		log.Log.Info("从path: ", routePath, " 获取路由数据为空")
		serviceZkData.Route = &common.ServiceRoute{RouteItem: []*common.RouteItem{}}
	} else {
		_, fData, err := common.DecodeValue(routeData)
		if err != nil {
			log.Log.Info("解析路由数据出错", err)
		}
		serviceZkData.Route = route.ParseRouteData(fData)
		//使用route进行过滤数据
		service.ProviderList = route.FilterServiceByRouteData(serviceZkData.Route, f.config.MeteData.Address, service.ProviderList)
	}

	return service, nil
}
