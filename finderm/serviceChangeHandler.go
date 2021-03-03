package cexport

import common "git.xfyun.cn/AIaaS/finder-go/common"

type serviceChangeHandler struct {
	cache *serviceFinderCache
}

func (s *serviceChangeHandler) OnServiceInstanceConfigChanged(name string, apiVersion string, addr string, config *common.ServiceInstanceConfig) bool {

}

func (s *serviceChangeHandler) OnServiceConfigChanged(name string, apiVersion string, config *common.ServiceConfig) bool {
	panic("implement me")
}

func (s *serviceChangeHandler) OnServiceInstanceChanged(name string, apiVersion string, eventList []*common.ServiceInstanceChangedEvent) bool {
	panic("implement me")
}

