package main

import (
	"finder-go/common"
	"fmt"
)

type ConfigChangedHandle struct {
}

func (s *ConfigChangedHandle) OnConfigFileChanged(config common.Config) bool {
	fmt.Println(config.Name, " has changed:\r\n", string(config.File))
	return true
}
