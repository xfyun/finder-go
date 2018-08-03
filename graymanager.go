package finder

import (
	"encoding/json"
	"strings"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	storageCommon "git.xfyun.cn/AIaaS/finder-go/storage/common"
)

const (
	grayNodePathPrefix = "/gray"
)

func ParseGrayConfigData(serverId string, data []byte) (string, bool) {
	//解析数据，解析出pushId和后面的数据
	_, fData, err := common.DecodeValue(data)
	if err != nil {
		return "", false
	}
	var grayConfigMaps []map[string]interface{}
	if err := json.Unmarshal(fData, &grayConfigMaps); err != nil {
		logger.Info("  [getGrayData] 使用json反序列化数据 ", fData, " 出错 ", err)
		return "", false
	}
	logger.Info(grayConfigMaps)
	//如何解析数据,会不会出现一个server在两个灰度组的情况
	for _, value := range grayConfigMaps {

		groupId := value["group_id"]
		serverStr := value["server_list"].([]interface{})[0].(string)
		serverList := strings.Split(serverStr, ",")
		for _, server := range serverList {
			if strings.Compare(server, serverId) == 0 {
				//比较之后等于0，则代表在该灰度组中，直接返回
				return groupId.(string), true
			}
		}
	}
	return "", false
}
func GetGrayConfigData(f *ConfigFinder, path string, callback storageCommon.ChangedCallback) (string, bool) {
	var serverId string = f.config.MeteData.Address
	if !strings.HasSuffix(path, grayNodePathPrefix) {
		path += grayNodePathPrefix
	}
	logger.Info(path)
	//节点不存在如何处理？
	var data []byte
	var err error
	if callback != nil {
		data, err = f.storageMgr.GetDataWithWatch(path, callback)
	} else {
		data, err = f.storageMgr.GetData(path)

	}
	if err != nil {
		logger.Info(" [getGrayData] 根据 path:", path, "获取数据出错：", err)
		return "", false
	}
	return ParseGrayConfigData(serverId, data)
}
