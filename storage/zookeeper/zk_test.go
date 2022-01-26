package zookeeper

import (
	"github.com/cooleric/go-zookeeper/zk"
	"sort"
	"testing"
	"time"
)

func TestNewWatcher(t *testing.T) {
	conn ,_,err := zk.Connect([]string{},5*time.Second)
	if err != nil{
		panic(err)
	}
	res, err := conn.Create("/test",[]byte("1.2.3.4"),EPHEMERAL,zk.WorldACL(zk.PermAll))
	if err != nil{
		panic(err)
	}
	sort.Slice()
}

func add(a, b, c int) int {
	return a + b + c
}

type addInput struct {
	a, b, c int
}

func add2(in addInput) int {
	return in.a + in.b + in.c
}
