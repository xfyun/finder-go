package finderm

import (
	"fmt"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	Init("http://10.1.87.69:6868","1.1.1.1:22334")
	serviceManager.SubscribeService("guiderAllService","gas","webgate-ws","webgate-ws","1.0.0")
	go func() {
		//time.Sleep(2*time.Second)
		fmt.Println(ListenService("guiderAllService","gas","webgate-ws","1.0.0",1))
		fmt.Println("changed")
	}()
	time.Sleep(3*time.Second)
	serviceManager.RegisterService("guiderAllService","gas","webgate-ws","1.0.0")
	serviceManager.RegisterServiceWithAddr("guiderAllService","gas","webgate-ws","1.0.0","1.1.1.1:5555")
	serviceManager.UnRegisterServiceWithAddr("guiderAllService","gas","webgate-ws","1.0.0","1.1.1.1:5555")

	//serviceManager.UnRegisterService("guiderAllService","gas","webgate-ws","1.0.0")

	addrs,err:=serviceManager.SubscribeService("guiderAllService","gas","xist-ed","webgate-ws","1.0.0")
	fmt.Println(addrs,err)


	file,err:=ListenFile("guiderAllService","gas","webgate-ws","1.1.1.9","schema_iat.json",1)
	if err != nil{
		panic(err)
	}
	fmt.Println("file changed:",file)
}

func TestLisq(t *testing.T) {
	l := newListener()
	go l.Listen("1",1)
	time.Sleep(10*time.Millisecond)

	go func() {
		for {
			fmt.Println(l.Listen("1",1))
		}
	}()


	go func() {
		for {
			fmt.Println(l.Listen("1",2))
		}
	}()

	l.Send("1",1)
	l.Send("1",2)
	l.Send("1",3)
	select {

	}
}

func TestGetFile(t *testing.T) {
	Init("http://10.1.87.69:6868","1.1.1.1:22334")
	serviceManager.SubscribeService("guiderAllService","gas","webgate-ws","webgate-ws","1.0.0")
	for{
		fmt.Println(ListenService("guiderAllService","gas","webgate-ws","1.0.0",1))

	}
}
