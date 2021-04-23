package finderm


var(
	configManager *configCenterManager
	serviceManager *serviceFinderManager
	inited = false
)

func initCheck(){
	if !inited{
		panic("config center in not init")
	}
}

func Init(companion string,myAddr string){
	configManager = newConfigCenterManager(companion)
	serviceManager = newServiceFinderManager(companion,myAddr)
	inited = true
}

func SubscribeService(project,group,myService,subScribeService,apiVersion string)([]string,error){
	initCheck()
	return serviceManager.SubscribeService(project,group,myService,subScribeService,apiVersion)
}

func RegisterService(project,group,service string,ver string)error{
	initCheck()
	return serviceManager.RegisterService(project,group,service,ver)
}

func RegisterServiceWithAddr(project,group,service string,ver ,addr string)error{
	initCheck()
	return serviceManager.RegisterServiceWithAddr(project,group,service,ver,addr)
}

func UnRegisterService(project,group,service string,ver string)error{
	initCheck()
	return serviceManager.UnRegisterService(project,group,service,ver)
}

func UnRegisterServiceWithAddr(project,group,service string,ver,addr string)error{
	initCheck()
	return serviceManager.UnRegisterServiceWithAddr(project,group,service,ver,addr)
}


func GetFile(project,group,service,version,file string)([]byte,error){
	initCheck()
	return configManager.GetFile(project,group,service,version,file)
}

func ListenService(project,group,service, apiVersion string, queue int)([]string,error){
	initCheck()
	addr,err:=serviceListener.Listen(assembleServiceListenerKey(project,group,service, apiVersion),int64(queue))
	if err != nil{
		return nil, err
	}
	return addr.([]string),nil
}



func ListenFile(project,group,service,version, file string, queue int)([]byte,error){
	initCheck()
	data,err:=configListener.Listen(assembleConfigListener(project,group,service, version,file),int64(queue))
	if err != nil{
		return nil, err
	}
	return data.([]byte),nil
}
