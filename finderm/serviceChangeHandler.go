package finderm

import (
	"fmt"
	common "git.xfyun.cn/AIaaS/finder-go/common"
	"git.xfyun.cn/AIaaS/finder-go/log"
)

type serviceChangeHandler struct {
	cache *serviceFinderCache

}

func (s *serviceChangeHandler) OnServiceInstanceConfigChanged(name string, apiVersion string, addr string, config *common.ServiceInstanceConfig) bool {

	return true
}

func (s *serviceChangeHandler) OnServiceConfigChanged(name string, apiVersion string, config *common.ServiceConfig) bool {
	return true
}

func (s *serviceChangeHandler) OnServiceInstanceChanged(name string, apiVersion string, eventList []*common.ServiceInstanceChangedEvent) bool {
	key:=assembleServiceFindKey(name,apiVersion)
	cachedAddrs,ok:=s.cache.serviceCache.Load(key)
	if !ok{
		return true
	}
	addrs:=cachedAddrs.([]string)
	for _, event := range eventList {
		if event.EventType == common.INSTANCEADDED{
			for _, instance := range event.ServerList {
				addrs = addToList(addrs,instance.Addr)
			}
		}
		if event.EventType == common.INSTANCEREMOVE{
			for _, instance := range event.ServerList {
				addrs = removeListE(addrs,instance.Addr)
			}
		}
	}
	sc:=s.cache
	s.cache.serviceCache.Store(key,addrs)
	lisKey:=assembleServiceListenerKey(sc.project,sc.group,name,apiVersion)
	if err:=serviceListener.Send(lisKey,addrs);err !=nil{
		//todo add log
		log.Println("service change send event error:",err)
	}

	return true
}

func assembleServiceListenerKey(project,group,service,version string)string{
	return fmt.Sprintf("%s.%s.%s.%s",project,group,service,version)
}

func addToList(list []string,val string)[]string{
	for _, s := range list {
		if s == val{
			return list
		}
	}
	return append(list,val)
}

func removeListE(list []string,val string)[]string{
	for i, s := range list {
		if s == val{
			return append(list[:i],list[i+1:]...)
		}
	}
	return list
}

