package finder

import (
	"encoding/json"
	"strings"

	common "git.xfyun.cn/AIaaS/finder-go/common"
	storageCommon "git.xfyun.cn/AIaaS/finder-go/storage/common"
	"log"
)

const (
	grayNodePathPrefix = "/gray"
)

func ParseGrayConfigData(serverId string, data []byte) (map[string]string, bool) {
	//解析数据，解析出pushId和后面的数据
	_, fData, err := common.DecodeValue(data)

	if err != nil {
		logger.Info("  [getGrayData] DecodeValue 出错 ", err)
		return nil, false
	}
	var grayConfigMaps []map[string]interface{}
	if err := json.Unmarshal(fData, &grayConfigMaps); err != nil {
		logger.Info("  [getGrayData] 使用json反序列化数据 ", fData, " 出错 ", err)
		return nil, false
	}
	//如何解析数据,会不会出现一个server在两个灰度组的情况
	garyConfig := make(map[string]string)
	for _, value := range grayConfigMaps {

		groupId := value["group_id"]
		serverStr := value["server_list"].([]interface{})[0].(string)
		serverList := strings.Split(serverStr, ",")
		for _, server := range serverList {
			garyConfig[server] = groupId.(string)
		}
	}
	return garyConfig, true

}
func GetGrayConfigData(f *ConfigFinder, path string, callback storageCommon.ChangedCallback) error {
	var serverId string = f.config.MeteData.Address
	if !strings.HasSuffix(path, grayNodePathPrefix) {
		path += grayNodePathPrefix
	}
	//节点不存在如何处理？
	var data []byte
	var err error
	if callback != nil {

		data, err = f.storageMgr.GetDataWithWatchV2(path, callback)
	} else {
		data, err = f.storageMgr.GetData(path)

	}
	if err != nil {
		if strings.Compare(err.Error(),common.ZK_NODE_DOSE_NOT_EXIST)==0{
			err:=f.storageMgr.SetPath(path)
			if err != nil {
				log.Println(" [getGrayData] 根据 path:", path, "创建节点出错：", err)
			}
			return nil
		}
		return err
	}
	if data==nil || len(data)==0{
		return nil
	}
	if grayConfig, ok := ParseGrayConfigData(serverId, data); ok {
		for key, value := range grayConfig {
			f.grayConfig.Store(key, value)
		}
		return nil
	}
	return nil
}
