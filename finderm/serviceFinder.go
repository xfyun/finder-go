package cexport

import (
	"git.xfyun.cn/AIaaS/finder-go"
	common "git.xfyun.cn/AIaaS/finder-go/common"
	"sync"
)

type serviceFinderCache struct {
	serviceCache sync.Map
	finder      *finder.FinderManager
	callBacks   sync.Map
}

func assembleServiceFindKey(service ,ver string)string{
	return service+"_"+ver
}

func (s *serviceFinderCache)SubscribeService(service string,apiVersion string)(addrs []string,err error){
	key:=assembleServiceFindKey(service,apiVersion)
	srvAddr,ok:=s.serviceCache.Load(key)
	if !ok{
		srv,err:=s.finder.ServiceFinder.UseAndSubscribeService([]common.ServiceSubscribeItem{{
			ServiceName: service,
			ApiVersion:  apiVersion,
		}},&serviceChangeHandler{})
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
