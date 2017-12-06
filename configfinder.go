package finder

import (
	"finder-go/common"
	"finder-go/errors"
	"finder-go/utils/zkutil"
	"sync"

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
			configFiles = append(configFiles, common.Config{Name: n, File: data})
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

	var data []byte
	configFiles := make([]common.Config, 0)
	for _, n := range name {
		data, err = f.zkManager.GetNodeData(f.zkManager.MetaData.ConfigRootPath + "/" + n)
		if err != nil {
			// todo
		} else {
			configFiles = append(configFiles, common.Config{Name: n, File: data})
			err = f.zkManager.GetNodeDataW(f.zkManager.MetaData.ConfigRootPath+"/"+n, func(c curator.CuratorFramework, e curator.CuratorEvent) error {
				//	configFiles = append(configFiles, common.Config{Name: n, File: data})
				return nil
			})
			if err != nil {
				// todo
			}
		}

		zkutil.ConfigEventPool.Append(n, event)
	}

	return configFiles, err
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

func (f *ConfigFinder) useAndSubscribeConfig2(name []string, callback AsyncConfigCallback, event zkutil.OnCfgUpdateEvent) error {
	var err error
	if len(name) == 0 {
		err = &errors.FinderError{
			Ret:  common.ConfigMissName,
			Func: "UseAndSubscribeConfig",
		}
		return err
	}

	wg := &sync.WaitGroup{}
	var data []byte
	// configFiles := make([]common.Config, 0)
	configFiles := new([]common.Config)
	for _, n := range name {
		wg.Add(1)

		err = f.zkManager.GetNodeDataW(f.zkManager.MetaData.ConfigRootPath+"/"+n, func(c curator.CuratorFramework, e curator.CuratorEvent) error {
			*configFiles = append(*configFiles, common.Config{Name: n, File: data})
			wg.Done()
			return nil
		})
		if err != nil {
			// todo
		}
		zkutil.ConfigEventPool.Append(n, event)
	}
	wg.Wait()
	callback(*configFiles)

	return err
}
