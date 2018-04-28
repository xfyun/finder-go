package zookeeper

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"git.xfyun.cn/AIaaS/finder-go/storage/common"
	"github.com/cooleric/go-zookeeper/zk"
)

type ZkManager struct {
	watcherPool map[string]common.DataChangedCallback
	conn        *zk.Conn
	params      map[string]string
	exit        chan bool
}

func NewZkManager(params map[string]string) (*ZkManager, error) {
	zm := &ZkManager{
		watcherPool: make(map[string]common.DataChangedCallback),
		params:      params,
	}

	return zm, nil
}

func (zm *ZkManager) Init() error {
	server_str, exist := zm.params["servers"]
	if !exist || len(server_str) == 0 {
		return errors.New("the param servers is empty")
	}
	servers := strings.Split(server_str, ",")

	timeout, exist := zm.params["session_timeout"]
	if !exist || len(timeout) == 0 {
		return errors.New("the param session_timeout is empty")
	}

	session_timeout, err := strconv.Atoi(timeout)
	if err != nil {
		return err
	}

	callback := zk.WithEventCallback(eventCallback)
	conn, _, err := zk.Connect(servers, time.Duration(session_timeout)*time.Millisecond, callback)
	if err != nil {
		return err
	}
	zm.conn = conn

	return nil
}

func (zm *ZkManager) Destroy() error {
	log.Println("exit send.")
	zm.params = nil
	zm.watcherPool = nil

	zm.conn.Close()
	log.Println("close end.")
	go func() {
		log.Println("send exit sigterm.")
		zm.exit <- true
	}()

	log.Println("destroied")
	return nil
}

func (zm *ZkManager) GetData(path string) ([]byte, error) {
	data, _, err := zm.conn.Get(path)
	return data, err
}

func (zm *ZkManager) Watch(path string) error {
	go func(m *ZkManager, p string) {
		for {
			_, _, event, err := zm.conn.GetW(path)
			if err != nil {
				log.Println(err)
				continue
			}
			select {
			case e, ok := <-event:
				if !ok {
					log.Println("<-event; !ok")
					return
				}
				log.Println(e.Type, "_", e.State, "_", e.Path)
			case exit, ok := <-m.exit:
				if !ok {
					log.Println("<-exit; !ok")
					return
				}
				if exit {
					log.Println("received exit sigterm.")
					return
				}
			}
		}
	}(zm, path)

	return nil
}

func (zm *ZkManager) GetChildren(path string) ([]string, error) {
	nodes, _, err := zm.conn.Children(path)
	return nodes, err
}

func (zm *ZkManager) GetChildrenWithWatch(path string) ([]string, error) {
	nodes, _, err := zm.conn.Children(path)
	return nodes, err
}

func (zm *ZkManager) SetPath(path string) error {
	_, err := zm.conn.Create(path, []byte{}, PERSISTENT, zk.WorldACL(zk.PermAll))
	if err == zk.ErrNoNode {
		err = makeDirs(zm.conn, path, false)
		if err != nil {
			_, err = zm.conn.Create(path, []byte{}, PERSISTENT, zk.WorldACL(zk.PermAll))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (zm *ZkManager) SetTempPath(path string) error {
	_, err := zm.conn.Create(path, []byte{}, EPHEMERAL, zk.WorldACL(zk.PermAll))
	if err == zk.ErrNoNode {
		err = makeDirs(zm.conn, path, false)
		if err != nil {
			_, err = zm.conn.Create(path, []byte{}, EPHEMERAL, zk.WorldACL(zk.PermAll))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (zm *ZkManager) SetData(path string, value []byte) error {
	_, err := zm.conn.Set(path, value, DEFAULT_VERSION)
	return err
}

func (zm *ZkManager) Remove(path string) error {
	return zm.conn.Delete(path, DEFAULT_VERSION)
}

func (zm *ZkManager) RemoveInRecursive(path string) error {
	return recursiveDelete(zm.conn, path, true)
}

func (zm *ZkManager) Watch2(path string, callback common.DataChangedCallback) error {

	return nil
}

func (zm *ZkManager) UnWatch(path string) error {
	return nil
}
