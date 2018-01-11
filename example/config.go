package main

import (
	"fmt"

	common "git.xfyun.cn/AIaaS/finder-go/common"
)

type ConfigChangedHandle struct {
}

func (s *ConfigChangedHandle) OnConfigFileChanged(config *common.Config) bool {
	fmt.Println(config.Name, " has changed:\r\n", string(config.File))
	return true
}
