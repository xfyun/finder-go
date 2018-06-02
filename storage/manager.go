package storage

import (
	"log"

	"git.xfyun.cn/AIaaS/finder-go/storage/zookeeper"
)

type StorageManager interface {
	Init() error
	Destroy() error
	GetData(path string) ([]byte, error)
	//GetDataWithWatch(path string) ([]byte, error)
	GetChildren(path string) ([]string, error)
	GetChildrenWithWatch(path string) ([]string, error)
	SetPath(path string) error
	SetTempPath(path string) error
	SetData(path string, value []byte) error
	Remove(path string) error
	RemoveInRecursive(path string) error
	//Watch(path string, callback common.DataChangedCallback) error
	Watch(path string) error
	UnWatch(path string) error
}

func NewManager(config *StorageConfig) (StorageManager, error) {
	switch config.Name {
	case "zookeeper":
		log.Println("called NewZkManager")
		return zookeeper.NewZkManager(config.Params)
	}
	return nil, nil
}
