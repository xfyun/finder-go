package finder

import (
	"finder-go/common"
	"finder-go/errors"
	"finder-go/utils/zkutil"

	"github.com/curator-go/curator"
)

var (
	configEventPrefix = "config_"
)

type AsyncConfigCallback func([]common.Config)

type ConfigFinder struct {
	config    *common.BootConfig
	zkManager *zkutil.ZkManager
}

func (f *ConfigFinder) UseConfig(name []string) ([]*common.Config, error) {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  errors.ConfigMissName,
			Func: "UseConfig",
		}

		return nil, err
	}
	configFiles := make([]*common.Config, 0)
	var data []byte
	for _, n := range name {
		data, err = f.zkManager.GetNodeData(f.zkManager.MetaData.ConfigRootPath + "/" + n)
		if err != nil {
			// todo
		} else {
			var fData []byte
			_, fData, err = common.DecodeValue(data)
			if err != nil {
				// todo
			} else {
				configFiles = append(configFiles, &common.Config{Name: n, File: fData})
			}
		}
	}

	return configFiles, err
}

func (f *ConfigFinder) UseAndSubscribeConfig(name []string, handler common.ConfigChangedHandler) ([]*common.Config, error) {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  errors.ConfigMissName,
			Func: "UseAndSubscribeConfig",
		}

		return nil, err
	}

	fileChan := make(chan *common.Config)
	for _, n := range name {
		err = f.zkManager.GetNodeDataW(f.zkManager.MetaData.ConfigRootPath+"/"+n, func(c curator.CuratorFramework, e curator.CuratorEvent) error {
			_, file, err := common.DecodeValue(e.Data())
			if err != nil {
				fileChan <- &common.Config{}
				return err
			}
			fileChan <- &common.Config{Name: e.Name(), File: file}
			return nil
		})
		if err != nil {
			// todo
			fileChan <- &common.Config{}
			continue
		}

		interHandle := ConfigHandle{ChangedHandler: handler, config: f.config}
		zkutil.ConfigEventPool.Append(common.ConfigEventPrefix+n, &interHandle)
	}

	return waitConfigResult(fileChan, len(name)), nil
}

func (f *ConfigFinder) UnSubscribeConfig(name string) error {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  errors.ConfigMissName,
			Func: "UnSubscribeConfig",
		}
		return err
	}

	zkutil.ConfigEventPool.Remove(name)

	return nil
}

func waitConfigResult(fileChan chan *common.Config, fileNum int) []*common.Config {
	configFiles := make([]*common.Config, 0)
	index := 0
	for {
		select {
		case c := <-fileChan:
			index++
			if len(c.Name) > 0 {
				configFiles = append(configFiles, c)
			}
			if index == fileNum {
				close(fileChan)
				return configFiles
			}
		}
	}
}
