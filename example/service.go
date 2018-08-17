package main

import (
	"fmt"

	common "git.xfyun.cn/AIaaS/finder-go/common"
)

// ServiceChangedHandle ServiceChangedHandle
type ServiceChangedHandle struct {
}

// OnServiceInstanceConfigChanged OnServiceInstanceConfigChanged
func (s *ServiceChangedHandle) OnServiceInstanceConfigChanged(name string, apiVersion string,instance string, config *common.ServiceInstanceConfig) bool {
	fmt.Println("服务实例配置信息更改开始，服务名：", name,"  版本号：",apiVersion,"  提供者实例为：",instance)
	fmt.Println("----当前配置为:  ", config.IsValid, "  ", config.UserConfig)
	fmt.Println("服务实例配置信息更改结束, 服务名：", name,"  版本号：",apiVersion,"  提供者实例为：",instance)
	config.IsValid=false
	config.UserConfig="aasasasasasasa"
	return true
}

// OnServiceConfigChanged OnServiceConfigChanged
func (s *ServiceChangedHandle) OnServiceConfigChanged(name string, apiVersion string,config *common.ServiceConfig) bool {
	fmt.Println("服务配置信息更改开始，服务名：", name,"  版本号：",apiVersion)
	fmt.Println("-----当前配置为: ", config.JsonConfig)
	fmt.Println("服务配置信息更改结束, 服务名：", name,"  版本号：",apiVersion)
	config.JsonConfig="zyssss"
	return true
}

// OnServiceInstanceChanged OnServiceInstanceChanged
func (s *ServiceChangedHandle) OnServiceInstanceChanged(name string, apiVersion string,eventList []*common.ServiceInstanceChangedEvent) bool {
	fmt.Println("服务实例变化通知开始, 服务名：", name,"  版本号：",apiVersion)
	for _, e := range eventList {

		for _, inst := range e.ServerList {
			if e.EventType == common.INSTANCEREMOVE {
				fmt.Println("----服务提供者节点减少事件 ：", e.ServerList)
				fmt.Println("-----------减少的服务提供者节点信息:  ", )
				fmt.Println("----------------------- 地址: ", inst.Addr)
				fmt.Println("----------------------- 是否有效: ", inst.Config.IsValid)
				fmt.Println("----------------------- 配置: ", inst.Config.UserConfig)

			} else {
				fmt.Println("----服务提供者节点增加事件 ：", e.ServerList)
				fmt.Println("-----------增加的服务提供者节点信息:  ", )
				fmt.Println("----------------------- 地址: ", inst.Addr)
				fmt.Println("----------------------- 是否有效: ", inst.Config.IsValid)
				fmt.Println("----------------------- 配置: ", inst.Config.UserConfig)

			}
			inst.Addr="zy_tet"
			inst.Config=&common.ServiceInstanceConfig{}
		}
	}

	fmt.Println("服务实例变化通知结束, 服务名：", name,"  版本号：",apiVersion)
	return true
}
