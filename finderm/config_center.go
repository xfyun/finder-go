package finderm

import (
	"errors"
	"fmt"
	"git.iflytek.com/AIaaS/finder-go/v3"
	common "git.iflytek.com/AIaaS/finder-go/v3/common"
	"sync"
)


type configCenter struct {
	configCache sync.Map
	finder      *finder.FinderManager
	callBacks   sync.Map
	project string
	group string
	service string
	version string
}

func newConfigCenter (project,group,service,version,companion string)(*configCenter,error){
	fd,err:=finder.NewFinderWithLogger(common.BootConfig{
		CompanionUrl:  companion,
		CachePath:     ".",
		CacheConfig:   true,
		CacheService:  false,
		ExpireTimeout: 0,
		MeteData:     &common.ServiceMeteData{
			Project: project,
			Group:   group,
			Service: service,
			Version: version,
			Address: "-",
		} ,
	},nil)
	if err != nil{
		return nil, fmt.Errorf("create finder error:%w",err)
	}

	return &configCenter{
		configCache: sync.Map{},
		finder:      fd,
		project: project,
		group: group,
		service: service,
		version: version,
	},nil
}

func (c *configCenter)subScribeAndGetFile(name string)([]byte,error){
	files,err:=c.finder.ConfigFinder.UseAndSubscribeConfig([]string{name},&configChangerHandler{
		cache: c,
	})
	if err != nil{
		return nil, err
	}
	file:=files[name]
	if file == nil{
		return nil, errors.New("file is nil")
	}
	return file.File,nil
}

func (c *configCenter)GetConfig(name string)([]byte,error){
	cfg,ok:=c.configCache.Load(name)
	if !ok{
		file,err:=c.subScribeAndGetFile(name)
		if err != nil{
			return nil, err
		}
		c.configCache.Store(name,file)
		return file,nil
	}
	return cfg.([]byte),nil
}

func (c *configCenter)GetConfigAndWatch(name string,cb FileChangedCallBack)([]byte,error){
	file,err:=c.GetConfig(name)
	if err != nil{
		return nil, err
	}
	c.callBacks.Store(name,cb)
	return file,nil
}

func assembleServiceKey(project,group,service,version string)string{
	return fmt.Sprintf("%s-%s-%s-%s",project,group,service,version)
}

type FileChangedCallBack func(name string,data []byte)bool

type configCenterManager struct {
	centers sync.Map
	companion string
}

func (c *configCenterManager)GetFile(project,group,service,version,file string)([]byte,error){
	return c.GetFileWithCallBack(project,group,service,version,file,nil)
}



func (c *configCenterManager)GetFileWithCallBack(project,group,service,version,file string,cb FileChangedCallBack)([]byte,error){
	centerKey:=assembleServiceKey(project,group,service,version)
	ctr,ok:=c.centers.Load(centerKey)
	if !ok{
		newCtr,err:=newConfigCenter(project,group,service,version,c.companion)
		if err != nil{
			return nil, fmt.Errorf("create center error:%w",err)
		}
		c.centers.Store(centerKey,newCtr)
		ctr = newCtr
	}
	ctrPtr:=ctr.(*configCenter)
	if cb == nil{
		return ctrPtr.GetConfig(file)
	}
	return ctrPtr.GetConfigAndWatch(file,cb)
}


func newConfigCenterManager(companion string)*configCenterManager{
	return &configCenterManager{
		centers:   sync.Map{},
		companion: companion,
	}
}

