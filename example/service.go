package main

import (
	"fmt"
	common "git.xfyun.cn/AIaaS/finder-go/common"
)

// ServiceChangedHandle ServiceChangedHandle
type ServiceChangedHandle struct {
}

// OnServiceInstanceConfigChanged OnServiceInstanceConfigChanged
func (s *ServiceChangedHandle) OnServiceInstanceConfigChanged(name string, apiVersion string, instance string, config *common.ServiceInstanceConfig) bool {

	config.IsValid = false
	config.UserConfig = "aasasasasasasa"
	config = nil
	return true
}

// OnServiceConfigChanged OnServiceConfigChanged
func (s *ServiceChangedHandle) OnServiceConfigChanged(name string, apiVersion string, config *common.ServiceConfig) bool {
	config.JsonConfig = "zyssss"
	config = nil
	return true
}

// OnServiceInstanceChanged OnServiceInstanceChanged
func (s *ServiceChangedHandle) OnServiceInstanceChanged(name string, apiVersion string, eventList []*common.ServiceInstanceChangedEvent) bool {
	for eventIndex, e := range eventList {
		for index, inst := range e.ServerList {
			if e.EventType == common.INSTANCEREMOVE {
				fmt.Println("-----------------------减少的服务提供者 地址: ", inst.Addr)
			} else {
				fmt.Println("----------------------增加的服务提供者 地址: ", inst.Addr)
			}
			e.ServerList[index].Addr = "zy_tet"
			e.ServerList[index].Config = &common.ServiceInstanceConfig{}
		}
		eventList[eventIndex] = nil
	}

	return true
}
