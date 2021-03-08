package finderm

import (
	"fmt"
	"git.xfyun.cn/AIaaS/finder-go"
	common "git.xfyun.cn/AIaaS/finder-go/common"
	"sync"
)

type serviceFinderCache struct {
	serviceCache sync.Map
	finder      *finder.FinderManager
	callBacks   sync.Map
	project string
	group string
	service string
}

func newServiceFinderCache(project,group,service,companion,addr string)(*serviceFinderCache,error){
	fd,err:=finder.NewFinderWithLogger(common.BootConfig{
		CompanionUrl:  companion,
		CachePath:     ".",
		CacheConfig:   true,
		CacheService:  false,
		ExpireTimeout: 0,
		MeteData:     &common.ServiceMeteData{
			Project: project,
			Group:   group,
			Service: service,
			//Version: version,
			Address: addr,
		} ,
	},nil)
	if err != nil{
		return nil, fmt.Errorf("create finder error:%w",err)
	}
	return &serviceFinderCache{
		serviceCache: sync.Map{},
		finder:       fd,
		callBacks:    sync.Map{},
		project:      project,
		group:        group,
		service:      service,
	},nil
}

func assembleServiceFindKey(service ,ver string)string{
	return service+"."+ver
}

func (s *serviceFinderCache)SubscribeService(service string,apiVersion string)(addrs []string,err error){
	key:=assembleServiceFindKey(service,apiVersion)
	srvAddr,ok:=s.serviceCache.Load(key)
	if !ok{
		srv,err:=s.finder.ServiceFinder.UseAndSubscribeService([]common.ServiceSubscribeItem{{
			ServiceName: service,
			ApiVersion:  apiVersion,
		}},&serviceChangeHandler{cache: s})
		if err !=nil{
			return nil, err
		}
		for _, val := range srv {
			if val.ServiceName == service && val.ApiVersion == apiVersion{
				for _, instance := range val.ProviderList {
					addrs = append(addrs,instance.Addr)
				}
			}
		}
		s.serviceCache.Store(key,addrs)
		return addrs, nil
	}
	return srvAddr.([]string),nil
}

func (s *serviceFinderCache)RegisterAddr(version string)error{

	return s.finder.ServiceFinder.RegisterService(version)
}

func (s *serviceFinderCache)UnRegisterAddr(version string)error{
	return s.finder.ServiceFinder.UnRegisterService(version)
}

type serviceFinderManager struct {
	serviceCaches sync.Map
	companion string
	myAddr string
}

func assembleFinderKey(project,group, myService string)string{
	return project+"."+group+"."+ myService
}


func (s *serviceFinderManager)SubscribeService(project,group,myService,subScribeService,apiVersion string)([]string,error){
	cache,err:=s.getFinderM(project,group,myService)
	if err != nil{
		return nil,err
	}
	return cache.SubscribeService(subScribeService,apiVersion)
}

func (s *serviceFinderManager)getFinderM(project,group, myService string)(*serviceFinderCache,error){
	key:=assembleFinderKey(project,group, myService)
	cache,ok:=s.serviceCaches.Load(key)
	if !ok{
		sfc,err:=newServiceFinderCache(project,group, myService,s.companion,s.myAddr)
		if err != nil{
			return  nil,err
		}
		s.serviceCaches.Store(key,sfc)
		return sfc,nil
	}
	return cache.(*serviceFinderCache),nil
}

func (s *serviceFinderManager)RegisterService(project,group,service string,ver string)error{
	cache,err:=s.getFinderM(project,group,service)
	if err != nil{
		return err
	}
	return  cache.RegisterAddr(ver)
}


func (s *serviceFinderManager)RegisterServiceWithAddr(project,group,service string,ver,addr string)error{
	cache,err:=s.getFinderM(project,group,service)
	if err != nil{
		return err
	}
	return  cache.finder.ServiceFinder.RegisterServiceWithAddr(addr,ver)
}

func (s *serviceFinderManager)UnRegisterService(project,group,service string,ver string)error{
	cache,err:=s.getFinderM(project,group,service)
	if err != nil{
		return err
	}
	return cache.UnRegisterAddr(ver)
}


func (s *serviceFinderManager)UnRegisterServiceWithAddr(project,group,service string,ver,addr string)error{
	cache,err:=s.getFinderM(project,group,service)
	if err != nil{
		return err
	}
	return cache.finder.ServiceFinder.UnRegisterServiceWithAddr(ver,addr)
}

func newServiceFinderManager(companion string,myAddr string)*serviceFinderManager{
	return &serviceFinderManager{
		serviceCaches: sync.Map{},
		companion:     companion,
		myAddr: myAddr,
	}
}
