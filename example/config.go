package main

import (
	"fmt"

	common "git.xfyun.cn/AIaaS/finder-go/common"
)

// ConfigChangedHandle ConfigChangedHandle
type ConfigChangedHandle struct {
}

// OnConfigFileChanged OnConfigFileChanged
func (s *ConfigChangedHandle) OnConfigFileChanged(config *common.Config) bool {
	fmt.Println(config.Name, " has changed:\r\n", string(config.File))
	return true
}
