package finder

const (
	Config_Success = iota // 0 加载配置成功
	Config_ReadFailure  // 1 读数据失败
	Config_LoadFailure  // 2 加载配置失败
)

const (
	Service_Success = iota // 0 服务发现模块函数调用成功
)