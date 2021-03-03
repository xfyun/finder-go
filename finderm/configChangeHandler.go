package cexport

import common "git.xfyun.cn/AIaaS/finder-go/common"

type configChangerHandler struct {
	cache *configCenter
}

func (c configChangerHandler) OnConfigFileChanged(config *common.Config) bool {
	c.cache.configCache.Store(config.Name, config.File)
	cb, ok := c.cache.callBacks.Load(config.Name)
	if ok {
		return cb.(FileChangedCallBack)(config.Name, config.File)
	}
	return true
}

func (c configChangerHandler) OnConfigFilesAdded(configs map[string]*common.Config) bool {
	return true
}

func (c configChangerHandler) OnConfigFilesRemoved(configNames []string) bool {
	return true
}

func (c configChangerHandler) OnError(errInfo common.ConfigErrInfo) {
	//panic("implement me")
}
