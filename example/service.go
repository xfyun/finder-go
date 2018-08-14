package main

import (
	"fmt"

	common "git.xfyun.cn/AIaaS/finder-go/common"
)

// ServiceChangedHandle ServiceChangedHandle
type ServiceChangedHandle struct {
}

// OnServiceInstanceConfigChanged OnServiceInstanceConfigChanged
func (s *ServiceChangedHandle) OnServiceInstanceConfigChanged(name string, instance string, config *common.ServiceInstanceConfig) bool {
	fmt.Println("服务实例配置信息更改开始:  ", name)
	fmt.Println("当前配置为:  ", config.IsValid, "  ", config.UserConfig)
	fmt.Println("服务实例配置信息更改结束:  ", name)
	return true
}

// OnServiceConfigChanged OnServiceConfigChanged
func (s *ServiceChangedHandle) OnServiceConfigChanged(name string, config *common.ServiceConfig) bool {
	fmt.Println(name, "服务配置信息更改开始：")
	fmt.Println("服务名为：", name, " 当前配置为: ", config.JsonConfig)

	fmt.Println("服务配置信息更改结束：")
	return true
}

// OnServiceInstanceChanged OnServiceInstanceChanged
func (s *ServiceChangedHandle) OnServiceInstanceChanged(name string, eventList []*common.ServiceInstanceChangedEvent) bool {
	fmt.Println("服务实例变化通知开始 :")
	for _, e := range eventList {
		if e.EventType == common.INSTANCEREMOVE {
			fmt.Println("服务提供者节点减少事件:", e.ServerList)
		} else {
			fmt.Println("服务提供者节点增加事件:", e.ServerList)
		}
		for _, inst := range e.ServerList {
			fmt.Println("服务提供者节点地址: addr:", inst.Addr)
		}
	}

	fmt.Println("服务实例变化通知结束 :")
	return true
}
