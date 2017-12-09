package errors

type ReturnCode int

const (
	ConfigSuccess     = iota // 0 获取配置成功
	ConfigReadFailure        // 1 读数据失败
	ConfigLoadFailure        // 2 加载配置失败
)

const (
	Success        ReturnCode = 0
	InvalidParam   ReturnCode = 10000
	ConfigMissName ReturnCode = 10100
	ZkMissRootPath ReturnCode = 10200 + iota
	ZkMissAddr
	ZkGetInfo
)

const (
	ServiceMissAddr ReturnCode = 10300 + iota
	ServiceMissName
)
