package main

/*

#include "config_center.h"
#include <stdlib.h>


 */
import "C"
import (
	"git.iflytek.com/AIaaS/finder-go/finderm"
	"unsafe"
)

func newServiceResult()*C.SubscribeServiceResult{
	return (*C.SubscribeServiceResult)(C.malloc(C.ulong(unsafe.Sizeof(C.SubscribeServiceResult{}))))
}

func newNode()*C.Node{
	return (*C.Node)(C.malloc(C.ulong(unsafe.Sizeof(C.Node{}))))
}

func valueOfAddr(addrs []string)*C.SubscribeServiceResult{
	res:=newServiceResult()
	head:=newNode()
	p:=head
	for _, addr := range addrs {
		p.addr = C.CString(addr)
		p.next = newNode()
		p = p.next
	}
	res.length = C.int(len(addrs))
	res.addrList = head
	res.code = 0
	return res
}

//export InitCenter
func InitCenter(companion ,myAddr *C.char){
	finderm.Init(C.GoString(companion),C.GoString(myAddr))
}


//export SubscribeService
func SubscribeService(project,group,myservice,service ,apiVersion *C.char)*C.SubscribeServiceResult{
	addrs,err:=finderm.SubscribeService(C.GoString(project),C.GoString(group),C.GoString(myservice),C.GoString(service),C.GoString(apiVersion))
	if err != nil{
		res:=newServiceResult()
		res.code = 10000
		res.info = C.CString(err.Error())
		return nil
	}


	return valueOfAddr(addrs)
}

//export RegisterService
func RegisterService(project,group,service ,version *C.char)C.CommonResult{
	err:=finderm.RegisterService(C.GoString(project),C.GoString(group),C.GoString(service),C.GoString(version))
	if err != nil{
		return C.CommonResult{code:10002,info:C.CString(err.Error())}
	}
	return C.CommonResult{}
}

//export RegisterServiceWithAddr
func RegisterServiceWithAddr(project,group,service ,version,addr *C.char)C.CommonResult{
	err:=finderm.RegisterServiceWithAddr(C.GoString(project),C.GoString(group),C.GoString(service),C.GoString(version),C.GoString(addr))
	if err != nil{
		return C.CommonResult{code:10002,info:C.CString(err.Error())}
	}
	return C.CommonResult{}
}

//export UnRegisterService
func UnRegisterService(project,group,service ,version *C.char)C.CommonResult{
	err:=finderm.UnRegisterService(C.GoString(project),C.GoString(group),C.GoString(service),C.GoString(version))
	if err != nil{
		return C.CommonResult{code:10003,info:C.CString(err.Error())}
	}
	return C.CommonResult{}
}

//export UnRegisterServiceWithAddr
func UnRegisterServiceWithAddr(project,group,service ,version,addr *C.char)C.CommonResult{
	err:=finderm.UnRegisterServiceWithAddr(C.GoString(project),C.GoString(group),C.GoString(service),C.GoString(version),C.GoString(addr))
	if err != nil{
		return C.CommonResult{code:10003,info:C.CString(err.Error())}
	}
	return C.CommonResult{}
}


//export SubscribeFile
func SubscribeFile(project,group,service ,version,file *C.char)C.SubscribeConfigResult{
	data,err:=finderm.GetFile(C.GoString(project),C.GoString(group),C.GoString(service),C.GoString(version),C.GoString(file))
	if err != nil{
		return C.SubscribeConfigResult{code:10000,info:C.CString(err.Error())}
	}
	return C.SubscribeConfigResult{data:C.CString(*(*string)(unsafe.Pointer(&data))),name:file}
}



//export ListenService
func ListenService(project,group,service ,apiVersion *C.char, queue C.int)*C.SubscribeServiceResult{
	addrs,err:=finderm.ListenService(C.GoString(project),C.GoString(group),C.GoString(service),C.GoString(apiVersion),int(queue))
	if err != nil{
		res:=newServiceResult()
		res.code = 10000
		res.info = C.CString(err.Error())
		return res
	}
	return valueOfAddr(addrs)
}

//export ListenFile
func ListenFile(project,group,service ,version,file *C.char, queue C.int)C.SubscribeConfigResult{
	data,err:=finderm.ListenFile(C.GoString(project),C.GoString(group),C.GoString(service),C.GoString(version),C.GoString(file),int(queue))
	if err != nil{
		return C.SubscribeConfigResult{code:10000,info:C.CString(err.Error())}
	}
	return C.SubscribeConfigResult{data:C.CString(*(*string)(unsafe.Pointer(&data))),name:file}
}



func main()  {

}
