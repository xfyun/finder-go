package main

import (
	"fmt"
	"git.xfyun.cn/AIaaS/finder-go"

	"log"
	"strings"

	 "git.xfyun.cn/AIaaS/finder-go/common"
)

// ConfigChangedHandle ConfigChangedHandle
type ConfigChangedHandle struct {
}

// OnConfigFileChanged OnConfigFileChanged
func (s *ConfigChangedHandle) OnConfigFileChanged(config *finder.Config) bool {
	if strings.HasSuffix(config.Name, ".toml") {
		fmt.Println(config.Name, " has changed:\r\n", string(config.File), " \r\n 解析后的map为 ：", config.ConfigMap)
	} else {
		fmt.Println(config.Name, " has changed:\r\n", string(config.File))
	}
	config.File = nil
	config.Name = ""
	config.ConfigMap = nil
	config = nil
	return true
}

func (s *ConfigChangedHandle) OnConfigFilesAdded(configs map[string]*finder.Config) bool {
	for _, config := range configs {
		if strings.HasSuffix(config.Name, ".toml") {
			fmt.Println(config.Name, " has changed:\r\n", string(config.File), " \r\n 解析后的map为 ：", config.ConfigMap)
		} else {
			fmt.Println(config.Name, " has changed:\r\n", string(config.File))
		}
		config.File = nil
		config.Name = ""
		config.ConfigMap = nil
		config = nil
		return true
	}

	return true
}

func (s *ConfigChangedHandle) OnConfigFilesRemoved(configNames []string) bool {
	for _, n := range configNames {
		fmt.Println(n, "has removed.")
	}

	return true
}

func (s *ConfigChangedHandle) OnError(errInfo finder.ConfigErrInfo) {
	log.Println("配置文件出错：", errInfo)
}
