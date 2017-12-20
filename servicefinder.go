package finder

import (
	"encoding/json"
	"finder-go/common"
	"finder-go/errors"
	"finder-go/utils/stringutil"
	"finder-go/utils/zkutil"
	"fmt"

	"github.com/curator-go/curator"
	"github.com/samuel/go-zookeeper/zk"
)

type ServiceFinder struct {
	zkManager *zkutil.ZkManager
	config    *common.BootConfig
}

func (f *ServiceFinder) RegisterService() error {
	var err error
	addr := f.config.MeteData.Address
	if stringutil.IsNullOrEmpty(addr) {
		err = &errors.FinderError{
			Ret:  errors.ServiceMissAddr,
			Func: "RegisterService",
		}

		return err
	}

	var data []byte
	data, err = getDefaultServiceItemConfig(addr)
	if err != nil {
		return err
	}
	parentPath := fmt.Sprintf("%s/%s/provider", f.zkManager.MetaData.ServiceRootPath, f.config.MeteData.Service)
	err = register(f.zkManager, parentPath, f.config.MeteData.Address, data)
	if err != nil {
		return err
	}
	err = registerService(f.config.CompanionUrl, f.config.MeteData.Project, f.config.MeteData.Group, f.config.MeteData.Service)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func (f *ServiceFinder) RegisterServiceWithAddr(addr string) error {
	var err error
	if stringutil.IsNullOrEmpty(addr) {
		err = &errors.FinderError{
			Ret:  errors.ServiceMissAddr,
			Func: "RegisterService",
		}

		return err
	}

	var data []byte
	data, err = getDefaultServiceItemConfig(addr)
	if err != nil {
		return err
	}
	parentPath := fmt.Sprintf("%s/%s/provider", f.zkManager.MetaData.ServiceRootPath, f.config.MeteData.Service)

	return register(f.zkManager, parentPath, addr, data)
}

func (f *ServiceFinder) UnRegisterService() error {
	servicePath := fmt.Sprintf("%s/%s/provider/%s", f.zkManager.MetaData.ServiceRootPath, f.config.MeteData.Service, f.config.MeteData.Address)

	return f.zkManager.RemoveInRecursive(servicePath)
}

func (f *ServiceFinder) UnRegisterServiceWithAddr(addr string) error {
	servicePath := fmt.Sprintf("%s/%s/provider/%s", f.zkManager.MetaData.ServiceRootPath, f.config.MeteData.Service, addr)

	return f.zkManager.RemoveInRecursive(servicePath)
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

	var addrList []string
	serviceList := make(map[string]*common.Service)
	for _, n := range name {
		servicePath := fmt.Sprintf("%s/%s/provider", f.zkManager.MetaData.ServiceRootPath, n)
		fmt.Println("useservice:", servicePath)
		addrList, err = f.zkManager.GetChildren(servicePath)
		if err != nil {
			fmt.Println("useservice:", err)
			service, err := GetServiceFromCache(f.config.CachePath, n)
			if err != nil {
				fmt.Println(err)
				//todo notify
			} else {
				serviceList[n] = service
			}
		} else if len(addrList) > 0 {
			fmt.Println("sp", servicePath)
			fmt.Println(addrList)
			serviceList[n] = getService(f.zkManager, servicePath, n, addrList)
			err = CacheService(f.config.CachePath, serviceList[n])
			if err != nil {
				fmt.Println("CacheService failed")
			}
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

	serviceChan := make(chan *common.Service)
	interHandle := &ServiceHandle{ChangedHandler: handler, config: f.config, zkManager: f.zkManager}
	for _, n := range name {
		servicePath := fmt.Sprintf("%s/%s/provider", f.zkManager.MetaData.ServiceRootPath, n)
		err = f.zkManager.GetChildrenW(servicePath, func(c curator.CuratorFramework, e curator.CuratorEvent) error {
			addrList := e.Children()
			if len(addrList) > 0 {
				service := getServiceWithWatcher(f.zkManager, servicePath, n, addrList, interHandle)
				if len(service.Name) > 0 {
					err = CacheService(f.config.CachePath, service)
					if err != nil {
						fmt.Println("CacheService failed")
					}
					serviceChan <- service
				} else {
					service, err := GetServiceFromCache(f.config.CachePath, n)
					if err != nil {
						fmt.Println(err)
						//todo notify
						serviceChan <- &common.Service{}
					} else {
						serviceChan <- service
					}
				}

				return nil
			}
			serviceChan <- &common.Service{}
			return nil
		})
		// handleChan := ServiceHandle{ChangedHandler: handler}
		if err != nil {
			service, err := GetServiceFromCache(f.config.CachePath, n)
			if err != nil {
				fmt.Println(err)
				//todo notify
				serviceChan <- &common.Service{}
			} else {
				serviceChan <- service
			}

			continue
		}

		zkutil.ServiceEventPool.Append(common.ServiceEventPrefix+n, interHandle)
	}

	return waitServiceResult(serviceChan, len(name)), nil
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

	zkutil.ServiceEventPool.Remove(name)

	return nil
}

func register(zm *zkutil.ZkManager, parentPath string, addr string, data []byte) error {
	var node *zk.Stat
	var err error
	servicePath := parentPath + "/" + addr
	node, err = zm.ExistsNode(servicePath)
	if err != nil {
		fmt.Println("ExistsNode", err)
		return err
	}
	if node == nil {
		err = createParentNode(zm, parentPath)
		if err != nil {
			fmt.Println("createParentNode", err)
			return err
		}

		return createTempNode(zm, servicePath, data)
	}

	return nil
}

func createParentNode(zm *zkutil.ZkManager, parentPath string) error {
	node, err := zm.ExistsNode(parentPath)
	if err != nil {
		return err
	}

	if node == nil {
		var result string
		result, err = zm.CreatePath(parentPath)
		if err != nil {
			return err
		}
		fmt.Println(result)
	}

	return nil
}

func createTempNode(zm *zkutil.ZkManager, path string, data []byte) error {
	result, err := zm.CreateTempPathWithData(path, data)
	if err != nil {
		return err
	}
	fmt.Println(result)

	return nil
}

func getDefaultServiceItemConfig(addr string) ([]byte, error) {
	defaultServiceInstanceConfig := common.ServiceInstanceConfig{
		Weight:  100,
		IsValid: true,
	}

	data, err := json.Marshal(defaultServiceInstanceConfig)
	if err != nil {
		return nil, err
	}

	var encodedData []byte
	encodedData, err = common.EncodeValue("", data)
	if err != nil {
		return nil, err
	}

	return encodedData, nil
}

func getServiceInstance(zm *zkutil.ZkManager, path string, addr string) (*common.ServiceInstance, error) {
	data, err := zm.GetNodeData(path + "/" + addr)
	if err != nil {
		return nil, err
	}

	var item []byte
	_, item, err = common.DecodeValue(data)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(item))
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

func getService(zm *zkutil.ZkManager, servicePath string, name string, addrList []string) *common.Service {
	var service = &common.Service{Name: name, ServerList: make([]*common.ServiceInstance, 0), Config: &common.ServiceConfig{}}
	for _, addr := range addrList {
		serviceInstance, err := getServiceInstance(zm, servicePath, addr)
		if err != nil {
			fmt.Println(err)
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

func getServiceWithWatcher(zm *zkutil.ZkManager, servicePath string, name string, addrList []string, interHandle *ServiceHandle) *common.Service {
	var service = &common.Service{Name: name, ServerList: make([]*common.ServiceInstance, 0), Config: &common.ServiceConfig{}}
	for _, addr := range addrList {
		serviceInstance, err := getServiceInstanceWithWatcher(zm, servicePath, addr, interHandle)
		if err != nil {
			fmt.Println(err)
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

func getServiceInstanceWithWatcher(zm *zkutil.ZkManager, servicePath string, addr string, interHandle *ServiceHandle) (*common.ServiceInstance, error) {
	serviceInstanceChan := make(chan *common.ServiceInstance)
	err := zm.GetNodeDataW(servicePath+"/"+addr, func(c curator.CuratorFramework, e curator.CuratorEvent) error {
		_, item, err := common.DecodeValue(e.Data())
		if err != nil {
			serviceInstanceChan <- &common.ServiceInstance{}
			return err
		}
		serviceInstance := &common.ServiceInstance{Addr: addr, Config: new(common.ServiceInstanceConfig)}
		err = json.Unmarshal(item, serviceInstance.Config)
		if err != nil {
			serviceInstanceChan <- &common.ServiceInstance{}
			return err
		}

		serviceInstanceChan <- serviceInstance
		return nil
	})
	if err != nil {
		return nil, err
	}
	zkutil.ServiceEventPool.Append(common.ServiceProviderEventPrefix+addr, interHandle)

	return waitServiceInstanceResult(serviceInstanceChan), nil
}

func waitServiceResult(serviceChan chan *common.Service, serviceNum int) map[string]*common.Service {
	serviceList := make(map[string]*common.Service)
	index := 0
	for {
		select {
		case s := <-serviceChan:
			index++
			if len(s.Name) > 0 {
				serviceList[s.Name] = s
			}
			if index == serviceNum {
				close(serviceChan)
				return serviceList
			}
		}
	}
}

func waitServiceInstanceResult(serviceInstanceChan chan *common.ServiceInstance) *common.ServiceInstance {
	serviceInstance := new(common.ServiceInstance)
	for {
		select {
		case s := <-serviceInstanceChan:
			if len(s.Addr) > 0 {
				serviceInstance = s
			}
			close(serviceInstanceChan)
			return serviceInstance
		}
	}
}
