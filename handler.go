package finder

import "finder-go/common"
import "fmt"

type ServiceHandle struct {
	ChangedHandler common.ServiceChangedHandler
}

func (s *ServiceHandle) OnServiceInstanceConfigChanged(name string, instance string, data []byte) {

}

func (s *ServiceHandle) OnServiceConfigChanged(name string, data []byte) {

}

func (s *ServiceHandle) OnServiceInstanceChanged(name string, instances []string) {

}

type ConfigHandle struct {
	ChangedHandler common.ConfigChangedHandler
}

func (s *ConfigHandle) OnConfigFileChanged(name string, data []byte) {
	pushID, file, err := common.DecodeValue(data)
	if err != nil {
		// todo
	} else {
		c := common.Config{
			Name: name,
			File: file,
		}

		ok := s.ChangedHandler.OnConfigFileChanged(c)
		if ok {

		}
		fmt.Println("load success:", pushID)
		// todo feedback
	}
}
