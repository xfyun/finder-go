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
	fmt.Println("收到 【", config.Name, " 】配置文件修改，最新内容如下 :\r\n", string(config.File))
	return true
}
