package finder

import (
	"encoding/json"
	"fmt"
	"log"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/utils/fileutil"
)

func CacheStorageInfo(cachePath string, zkInfo *common.StorageInfo) error {
	cachePath = fmt.Sprintf("%s/storage_%s.findercache", cachePath, "info")
	data, err := json.Marshal(zkInfo)
	if err != nil {
		log.Println(err)
		return err
	}
	err = fileutil.WriteFile(cachePath, data)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetStorageInfoFromCache(cachePath string) (*common.StorageInfo, error) {
	cachePath = fmt.Sprintf("%s/storage_%s.findercache", cachePath, "info")
	data, err := fileutil.ReadFile(cachePath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	zkInfo := &common.StorageInfo{}
	err = json.Unmarshal(data, zkInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return zkInfo, nil
}

func CacheConfig(cachePath string, config *common.Config) error {
	cachePath = fmt.Sprintf("%s/config_%s.findercache", cachePath, config.Name)

	var err error
	if fileutil.IsTomlFile(config.Name) {
		//如果是toml文件，则直接存入解析后的数据
		data, _ := json.Marshal(config.ConfigMap)
		err = fileutil.WriteFile(cachePath, data)
	} else {
		//如果是普通文件，则写入文件数据
		err = fileutil.WriteFile(cachePath, config.File)
	}
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetConfigFromCache(cachePath string, name string) (*common.Config, error) {
	cachePath = fmt.Sprintf("%s/config_%s.findercache", cachePath, name)

	exist, err := fileutil.ExistPath(cachePath)
	if err != nil {
		log.Println(err)
	}
	if !exist {
		err = errors.NewFinderError(errors.ConfigMissCacheFile)
		log.Printf(cachePath, err)
		return nil, err
	}
	data, err := fileutil.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}
	if fileutil.IsTomlFile(name) {
		//如果是toml文件
		var tomlConfig = make(map[string]interface{})
		err = json.Unmarshal(data, &tomlConfig)
		if err != nil {
			return nil, err
		}
		return &common.Config{Name: name, ConfigMap: tomlConfig}, nil
	} else {
		return &common.Config{Name: name, File: data}, nil
	}
}

func CacheService(cachePath string, service *common.Service) error {
	cachePath = fmt.Sprintf("%s/service_%s_%s.findercache", cachePath, service.ServiceName, service.ApiVersion)
	data, err := json.Marshal(service)
	if err != nil {
		log.Println(err)
		return err
	}
	err = fileutil.WriteFile(cachePath, data)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetServiceFromCache(cachePath string, item common.ServiceSubscribeItem) (*common.Service, error) {
	cachePath = fmt.Sprintf("%s/service_%s_%s.findercache", cachePath, item.ServiceName, item.ApiVersion)
	data, err := fileutil.ReadFile(cachePath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	service := &common.Service{}
	err = json.Unmarshal(data, service)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return service, nil
}
