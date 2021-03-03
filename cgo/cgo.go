package main

/*
#cgo  LDFLAGS: -ldl
#include "config_center.h"
#include <stdlib.h>

void ss(){
   return ;
}
 */
import "C"
import "git.xfyun.cn/AIaaS/finder-go/finderm"

//export SubscribeService
func SubscribeService(project,group,myservice,service ,apiVersion *C.char)*C.SubscribeServiceResult{

	addrs,err:=finderm.SubscribeService(C.GoString(project),C.GoString(group),C.GoString(myservice),C.GoString(service),C.GoString(apiVersion))
	if err != nil{
		res:=&C.SubscribeServiceResult{}
		res.code = 10000
		res.info = C.CString(err.Error())
		return nil
	}

	res:=&C.SubscribeServiceResult{}
	head:=&C.Node{}
	p:=head
	for _, addr := range addrs {
		p.addr = C.CString(addr)
		p.next = &C.Node{}
		p = p.next
	}
	res.length = C.int(len(addrs))
	res.addrList = head
	res.code = 0
	return res
}

//export Init
func Init(companion ,myAddr *C.char){
	finderm.Init(C.GoString(companion),C.GoString(myAddr))
}

func main(){

}
