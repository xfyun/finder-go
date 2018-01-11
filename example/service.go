package main

import (
	"fmt"

	common "git.xfyun.cn/AIaas/finder-go/common"
)

type ServiceChangedHandle struct {
}

func (s *ServiceChangedHandle) OnServiceInstanceConfigChanged(name string, instance string, config *common.ServiceInstanceConfig) bool {
	fmt.Println(name, " update begin:")
	fmt.Println("name:", name)
	fmt.Println("addr:", instance)
	fmt.Println("weight:", config.Weight)
	fmt.Println("is_valid:", config.IsValid)

	fmt.Println("got service update finish.")
	return true
}

func (s *ServiceChangedHandle) OnServiceConfigChanged(name string, config *common.ServiceConfig) bool {
	fmt.Println(name, " update begin:")
	fmt.Println("name:", name)
	fmt.Println("lb_mode:", config.LoadBalanceMode)
	fmt.Println("proxy_mode:", config.ProxyMode)

	fmt.Println("got service update finish.")
	return true
}

func (s *ServiceChangedHandle) OnServiceInstanceChanged(name string, eventList []*common.ServiceInstanceChangedEvent) bool {
	fmt.Println(name, " update begin:")
	for _, e := range eventList {
		fmt.Println("event:", e)
		for _, inst := range e.ServerList {
			fmt.Println("addr:", inst.Addr)
			fmt.Println("weight:", inst.Config.Weight)
			fmt.Println("is_valid:", inst.Config.IsValid)
		}
	}

	fmt.Println("got service update finish.")
	return true
}
