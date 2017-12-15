package finder

import (
	"encoding/json"
	"finder-go/common"
	"finder-go/utils/fileutil"
	"fmt"
)

func CacheZkInfo(cachePath string, zkInfo *common.ZkInfo) error {
	cachePath = fmt.Sprintf("%s/zk_%s.findercache", cachePath, "info")
	data, err := json.Marshal(zkInfo)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = fileutil.WriteFile(cachePath, data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func GetZkInfoFromCache(cachePath string) (*common.ZkInfo, error) {
	cachePath = fmt.Sprintf("%s/zk_%s.findercache", cachePath, "info")
	data, err := fileutil.ReadFile(cachePath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	zkInfo := &common.ZkInfo{}
	err = json.Unmarshal(data, zkInfo)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return zkInfo, nil
}

func CacheConfig(cachePath string, config *common.Config) error {
	cachePath = fmt.Sprintf("%s/config_%s.findercache", cachePath, config.Name)
	err := fileutil.WriteFile(cachePath, config.File)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func GetConfigFromCache(cachePath string, name string) ([]byte, error) {
	cachePath = fmt.Sprintf("%s/config_%s.findercache", cachePath, name)
	return fileutil.ReadFile(cachePath)
}

func CacheService(cachePath string, service *common.Service) error {
	cachePath = fmt.Sprintf("%s/service_%s.findercache", cachePath, service.Name)
	data, err := json.Marshal(service)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = fileutil.WriteFile(cachePath, data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func GetServiceFromCache(cachePath string, name string) (*common.Service, error) {
	cachePath = fmt.Sprintf("%s/service_%s.findercache", cachePath, name)
	data, err := fileutil.ReadFile(cachePath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	service := &common.Service{}
	err = json.Unmarshal(data, service)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return service, nil
}
