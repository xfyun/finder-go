package storage

import (


	"git.xfyun.cn/AIaaS/finder-go/storage/common"
	"git.xfyun.cn/AIaaS/finder-go/storage/zookeeper"
	"git.xfyun.cn/AIaaS/finder-go/log"
)

type StorageManager interface {
	GetZkNodePath()(string ,error)
	Init() error
	Destroy() error
	GetData(path string) ([]byte, error)
	GetDataWithWatchV2(path string, callback common.ChangedCallback) ([]byte, error)
	GetDataWithWatch(path string, callback common.ChangedCallback) ([]byte, error)
	GetChildren(path string) ([]string, error)
	GetChildrenWithWatch(path string, callback common.ChangedCallback) ([]string, error)
	SetPath(path string) error
	SetPathWithData(path string, data []byte) error
	SetTempPath(path string) error
	SetTempPathWithData(path string, data []byte) error
	SetData(path string, value []byte) error
	Remove(path string) error
	RemoveInRecursive(path string) error
	UnWatch(path string) error
	CheckExists(path string) (bool, error)
}

func NewManager(config *StorageConfig) (StorageManager, error) {
	switch config.Name {
	case "zookeeper":
		log.Log.Debug("called NewZkManager")
		return zookeeper.NewZkManager(config.Params)
	}
	return nil, nil
}
