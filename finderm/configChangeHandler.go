package finderm

import (
	"fmt"
	common "github.com/xfyun/finder-go/common"
	"log"
)

type configChangerHandler struct {
	cache *configCenter
}

func (c configChangerHandler) OnConfigFileChanged(config *common.Config) bool {
	c.cache.configCache.Store(config.Name, config.File)
	cc := c.cache
	if err := configListener.Send(assembleConfigListener(cc.project, cc.group, cc.service, cc.version, config.Name), config.File); err != nil {
		log.Println("send config change event error", err)
	}
	cb, ok := c.cache.callBacks.Load(config.Name)
	if ok {
		return cb.(FileChangedCallBack)(config.Name, config.File)
	}

	return true
}

func assembleConfigListener(project, group, service, version, file string) string {
	return fmt.Sprintf("%s.%s.%s.%s.%s", project, group, service, version, file)
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
