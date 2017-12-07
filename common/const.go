package common

const DefaultCacheDir = "findercache"

type ReturnCode int

const (
	ConfigSuccess     = iota // 0 获取配置成功
	ConfigReadFailure        // 1 读数据失败
	ConfigLoadFailure        // 2 加载配置失败
)

const (
	Success        ReturnCode = 0
	InvalidParam   ReturnCode = 10000 + iota
	ConfigMissName ReturnCode = 10100 + iota
	ZkMissRootPath ReturnCode = 10200 + iota
	ZkMissAddr
	ZkGetInfo
)

const (
	ServiceSuccess = iota // 0 服务发现模块函数调用成功
)
