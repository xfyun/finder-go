package finder

type ReturnCode int

func (retCode ReturnCode) String() string {
	if errString, ok := retCodeToString[retCode]; ok {
		return errString
	}
	return "unknow error for " + string(retCode)
}

var retCodeToString = map[ReturnCode]string{
	Success:                     "成功",
	InvalidParam:                "无效的参数",
	MissCompanionUrl:            "缺少companionUrl",
	ConfigMissName:              "[config] 配置信息中缺少配置的名字",
	ConfigMissCacheFile:         "[config] 丢失缓存文件",
	ZkGetInfoError:              "获取zk信息错误",
	ZkInfoMissConfigRootPath:    "获取的zkInfo中不存在config path",
	ZkInfoMissServiceRootPath:   "获取的zkInfo中不存在server path",
	ZkInfoMissAddr:              "获取的zkInfo中不存在zk的地址信息",
	ZkInfoAddrConvertError:      "转换获取的zkInfo中的zk地址信息出错",
	ZkParamsMissServers:         "zk参数中缺少zk的服务地址信息",
	ZkParamsMissSessionTimeout:  "zk参数中缺少sessionTimeout的配置信息",
	ZkDataCanotNil:              "zk中不能设置nil的数据",
	CompanionRegisterServiceErr: "向companion注册服务失败",
	FeedbackServiceError:        "feedback service数据到companion失败",
	FeedbackConfigError:         "feedback config数据到companion失败",
	DecodeVauleDataEmptyErr:     "解码数据的时候，数据为空",
	DecodeVauleDataNotFullErr:   "解码数据的时候，数据不完整",
	DecodeVauleDataFormatErr:    "解码数据的时候，数据格式出错",
	ServiceMissName:             "[service] 没有service的名字信息",
	ServiceMissAddr:             "[service] 缺失service对应的地址信息",
}

const (
	ConfigSuccess     = iota // 0 获取配置成功
	ConfigReadFailure        // 1 读数据失败
	ConfigLoadFailure        // 2 加载配置失败
)

const (
	Success          ReturnCode = 0
	InvalidParam     ReturnCode = 10000
	MissCompanionUrl ReturnCode = 10001
)

//config相关错误
const (
	ConfigMissName ReturnCode = 10100 + iota
	ConfigMissCacheFile
)

//zk相关错误
const (
	ZkGetInfoError ReturnCode = 10200 + iota
	ZkInfoMissRootPath
	ZkInfoMissConfigRootPath
	ZkInfoMissServiceRootPath
	ZkInfoMissAddr
	ZkInfoAddrConvertError

	ZkParamsMissServers
	ZkParamsMissSessionTimeout
	ZkDataCanotNil
)

//service相关错误
const (
	ServiceMissAddr ReturnCode = 10300 + iota
	ServiceMissName
)

//feedback相关错误
const (
	FeedbackConfigError ReturnCode = 10400 + iota
	FeedbackServiceError
)

//Companion相关错误
const (
	CompanionRegisterServiceErr ReturnCode = 10500 + iota
)

const (
	DecodeVauleDataEmptyErr ReturnCode = 10600 + iota
	DecodeVauleDataNotFullErr
	DecodeVauleDataFormatErr
)
