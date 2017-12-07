package finder

import (
	"finder-go/common"
	"finder-go/errors"
	"finder-go/utils/zkutil"

	"github.com/curator-go/curator"
)

type AsyncConfigCallback func([]common.Config)

type ConfigFinder struct {
	zkManager *zkutil.ZkManager
}

func (f *ConfigFinder) UseConfig(name []string) ([]common.Config, error) {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  common.ConfigMissName,
			Func: "UseConfig",
		}

		return nil, err
	}
	configFiles := make([]common.Config, len(name))
	var data []byte
	for _, n := range name {
		data, err = f.zkManager.GetNodeData(f.zkManager.MetaData.ConfigRootPath + "/" + n)
		if err != nil {
			// todo
		} else {
			var fData []byte
			_, fData, err = zkutil.DecodeValue(data)
			if err != nil {
				// todo
			} else {
				configFiles = append(configFiles, common.Config{Name: n, File: fData})
			}
		}
	}

	return configFiles, err
}

func (f *ConfigFinder) UseAndSubscribeConfig(name []string, event zkutil.OnCfgUpdateEvent) ([]common.Config, error) {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  common.ConfigMissName,
			Func: "UseAndSubscribeConfig",
		}

		return nil, err
	}

	fileChan := make(chan *common.Config)
	for _, n := range name {
		err = f.zkManager.GetNodeDataW(f.zkManager.MetaData.ConfigRootPath+"/"+n, func(c curator.CuratorFramework, e curator.CuratorEvent) error {
			pushId, fData, err := zkutil.DecodeValue(e.Data())
			if err != nil {
				fileChan <- &common.Config{}
				return err
			}
			fileChan <- &common.Config{PushId: pushId, Name: e.Name(), File: fData}
			return nil
		})
		if err != nil {
			// todo
			fileChan <- &common.Config{}
			continue
		}

		zkutil.ConfigEventPool.Append(n, event)
	}

	return waitResult(fileChan, len(name)), nil
}

func (f *ConfigFinder) UnSubscribeConfig(name string) error {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  common.ConfigMissName,
			Func: "UnSubscribeConfig",
		}
		return err
	}

	zkutil.ConfigEventPool.Remove(name)

	return nil
}

func (f *ConfigFinder) useAndSubscribeConfig(name []string, event zkutil.OnCfgUpdateEvent) ([]common.Config, error) {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  common.ConfigMissName,
			Func: "UseAndSubscribeConfig",
		}
		return nil, err
	}

	var data []byte
	configFiles := make([]common.Config, 0)
	for _, n := range name {
		data, err = f.zkManager.GetNodeData(f.zkManager.MetaData.ConfigRootPath + "/" + n)
		if err != nil {
			// todo
		} else {
			var fData []byte
			var pushId string
			pushId, fData, err = zkutil.DecodeValue(data)
			if err != nil {
				// todo
			} else {
				configFiles = append(configFiles, common.Config{PushId: pushId, Name: n, File: fData})
				err = f.zkManager.GetNodeDataW(f.zkManager.MetaData.ConfigRootPath+"/"+n, func(c curator.CuratorFramework, e curator.CuratorEvent) error {
					//	configFiles = append(configFiles, common.Config{Name: n, File: data})
					return nil
				})
				if err != nil {
					// todo
				}
			}

		}

		zkutil.ConfigEventPool.Append(n, event)
	}

	return configFiles, err
}

func waitResult(fileChan chan *common.Config, fileNum int) []common.Config {
	configFiles := make([]common.Config, 0)
	index := 0
	for {
		select {
		case c := <-fileChan:
			index++
			if len(c.Name) > 0 {
				configFiles = append(configFiles, *c)
			}
			if index == fileNum {
				close(fileChan)
				return configFiles
			}
		}
	}
}
