package common

const DefaultCacheDir = "findercache"

type ReturnCode int

const (
	ConfigSuccess     = iota // 0 加载配置成功
	ConfigReadFailure        // 1 读数据失败
	ConfigLoadFailure        // 2 加载配置失败
)

const (
	Success        ReturnCode = 0
	ConfigMissName ReturnCode = 10000 + iota
	ZkMissRootPath ReturnCode = 10000
	ZkMissAddr
)

const (
	ServiceSuccess = iota // 0 服务发现模块函数调用成功
)
