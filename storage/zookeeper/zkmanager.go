package zookeeper

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	errors "git.xfyun.cn/AIaaS/finder-go/errors"
	"git.xfyun.cn/AIaaS/finder-go/storage/common"
	"github.com/cooleric/go-zookeeper/zk"
)

type ZkManager struct {
	conn   *zk.Conn
	params map[string]string
	exit   chan bool
	//记录了临时路径
	tempPaths sync.Map
	//记录了path对应的Watcher
	dataWatcher sync.Map
}

func NewZkManager(params map[string]string) (*ZkManager, error) {
	zm := &ZkManager{
		tempPaths:   sync.Map{},
		params:      params,
		dataWatcher: sync.Map{},
	}
	return zm, nil
}

func (zm *ZkManager) GetZkNodePath() (string, error) {
	if path, ok := zm.params["zk_node_path"]; ok {
		return path, nil
	} else {
		return "", errors.NewFinderError(errors.ZkInfoMissZkNodePath)
	}
}
func (zm *ZkManager) Init() error {

	//必要参数判断
	serverStr, exist := zm.params["servers"]
	if !exist || len(serverStr) == 0 {
		return errors.NewFinderError(errors.ZkParamsMissServers)
	}

	servers := strings.Split(serverStr, ",")
	timeout, exist := zm.params["session_timeout"]
	if !exist || len(timeout) == 0 {
		return errors.NewFinderError(errors.ZkParamsMissSessionTimeout)
	}
	sessionTimeout, err := strconv.Atoi(timeout)
	if err != nil {
		return err
	}

	//新建zookeeper连接
	conn, _, err := zk.Connect(servers, time.Duration(sessionTimeout)*time.Millisecond, zk.WithEventCallback(zm.eventCallback))
	if err != nil {
		return err
	}

	zm.conn = conn
	return nil
}
func (zm *ZkManager) eventCallback(e zk.Event) {
	switch e.Type {
	case zk.EventSession:
		switch e.State {
		case zk.StateDisconnected:
			return
		case zk.StateConnecting:
			return
		case zk.StateConnected:
			return
		case zk.StateHasSession:
			go zm.recoverTempPaths()
			return
		case zk.StateExpired:
			return
		case zk.StateAuthFailed:
			return
		case zk.StateConnectedReadOnly:
			return
		case zk.StateSaslAuthenticated:
			return
		case zk.StateUnknown:
			return
		}
		return
	}
}

/**
 * 在恢复会话的时候进行调用，用来恢复临时路径
 */
func (zm *ZkManager) recoverTempPaths() {
	var err error
	zm.tempPaths.Range(func(key, value interface{}) bool {
		if value == nil {
			//如果path上的数据为空。则直接设置path
			for {
				err = zm.SetTempPath(key.(string))
				if err != nil {
					log.Println("caught an error:zm.SetTempPath in recoverTempPaths:", err)
					continue
				}
				break
			}
		} else {
			//如果path上的数据不为空。则直接设置path和对应的数据
			for {
				err = zm.SetTempPathWithData(key.(string), value.([]byte))
				if err != nil {
					log.Println("caught an error:zm.SetTempPathWithData in recoverTempPaths:", err)
					continue
				}

				break
			}
		}

		return true
	})
}

func (zm *ZkManager) Destroy() error {
	log.Println("exit send.")
	zm.params = nil

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
	//获取节点数据
	data, _, err := zm.conn.Get(path)
	return data, err
}

func (zm *ZkManager) GetDataWithWatchV2(path string, callback common.ChangedCallback) ([]byte, error) {

	//获取数据，并注册Watch
	data, _, event, err := zm.conn.GetW(path)
	if err != nil {
		return nil, err
	}

	//监听是否有watch到达
	go watchEvent(zm, event, callback)

	return data, err
}

func watchEvent(zm *ZkManager, event <-chan zk.Event, callback common.ChangedCallback) {
	select {
	case e, ok := <-event:
		if !ok {
			log.Println("<-event; !ok")
			return
		}
		log.Println("---------------GetDataWithWatchV2: ", e)
		callback.Process(e.Path, getNodeFromPath(e.Path))
		break
	case exit, ok := <-zm.exit:
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
func (zm *ZkManager) GetDataWithWatch(path string, callback common.ChangedCallback) ([]byte, error) {

	data, _, event, err := zm.conn.GetW(path)
	if err != nil {
		log.Println("[ GetDataWithWatch ]获取数据出错", err)
	}
	//返回的event
	go func(zm *ZkManager, p string, event <-chan zk.Event) {
		for {
			select {
			case e, ok := <-event:
				if !ok {
					log.Println("路径是: ", path, " 回调有误  ", e)
				}
				log.Println("收到通知，", e)
				if e.Type == zk.EventNodeDeleted {
					log.Println("节点删除事件，不再获取该节点的数据 ", e)
					return
				}
				var retryCount int32
				for {
					// 这个地方有问题，如果节点被删除的话，会成为死循环，修改为尝试三次

					data, _, event, err = zm.conn.GetW(path)
					if err != nil {
						log.Println("[ zkWatcher] 从", path, "获取数据失败 ", err)
						retryCount++
						if retryCount > 3 {
							time.Sleep(1 * time.Second)
							break
						}
						continue
					} else {
						callback.DataChangedCallback(e.Path, getNodeFromPath(e.Path), data)
					}

					break
				}
			case exit, ok := <-zm.exit:
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
	}(zm, path, event)

	return data, err
}

func (zm *ZkManager) GetChildren(path string) ([]string, error) {
	nodes, _, err := zm.conn.Children(path)
	return nodes, err
}

func (zm *ZkManager) GetChildrenWithWatch(path string, callback common.ChangedCallback) ([]string, error) {
	data, _, event, err := zm.conn.ChildrenW(path)
	if err != nil {

		return nil, err
	}

	go func(zm *ZkManager, p string, event <-chan zk.Event) {
		for {
			select {
			case e, ok := <-event:
				if !ok {
					log.Println("[ GetChildrenWithWatch ]  <-event; !ok")
					continue
				}
				for {
					data, _, event, err = zm.conn.ChildrenW(path)
					if err != nil {
						log.Println("[ GetChildrenWithWatch ] 再次获取字节点信息出错 ", err)
						continue
					} else {
						callback.ChildrenChangedCallback(e.Path, getNodeFromPath(e.Path), data)
					}

					break
				}
			case exit, ok := <-zm.exit:
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
	}(zm, path, event)

	return data, err
}

func (zm *ZkManager) SetPath(path string) error {
	return zm.SetPathWithData(path, []byte{})
}
func (zm *ZkManager) CheckExists(path string) (bool, error) {
	exists, _, err := zm.conn.Exists(path)
	if err != nil {
		return false, err
	}
	return exists, nil
}
func (zm *ZkManager) SetPathWithData(path string, data []byte) error {
	if data == nil {
		return errors.NewFinderError(errors.ZkDataCanotNil)
	}
	_, err := zm.conn.Create(path, data, PERSISTENT, zk.WorldACL(zk.PermAll))
	if err == zk.ErrNoNode {
		err = makeDirs(zm.conn, path, false)
		if err != nil {
			return err
		}
		_, err = zm.conn.Create(path, data, PERSISTENT, zk.WorldACL(zk.PermAll))
		if err != nil {
			return err
		}
	}

	return nil
}

func (zm *ZkManager) SetTempPath(path string) error {
	err := zm.SetTempPathWithData(path, []byte{})

	if err == nil {
		zm.tempPaths.Store(path, nil)
	}

	return err
}

func (zm *ZkManager) SetTempPathWithData(path string, data []byte) error {
	if data == nil {
		return errors.NewFinderError(errors.ZkDataCanotNil)
	}
	_, err := zm.conn.Create(path, data, EPHEMERAL, zk.WorldACL(zk.PermAll))
	if err == zk.ErrNoNode {
		err = makeDirs(zm.conn, path, false)
		if err != nil {
			return err
		}
		_, err = zm.conn.Create(path, data, EPHEMERAL, zk.WorldACL(zk.PermAll))
		if err != nil {
			return err
		}
	} else if err == zk.ErrNodeExists {
		err = zm.RemoveInRecursive(path)
		if err != nil {
			return err
		}
		_, err = zm.conn.Create(path, data, EPHEMERAL, zk.WorldACL(zk.PermAll))
		if err != nil {
			return err
		}
	} else {
		return err
	}

	zm.tempPaths.Store(path, data)

	return nil
}

func (zm *ZkManager) SetData(path string, value []byte) error {
	if value == nil {
		return errors.NewFinderError(errors.ZkDataCanotNil)
	}
	_, err := zm.conn.Set(path, value, DEFAULT_VERSION)
	return err
}

func (zm *ZkManager) Remove(path string) error {
	return zm.conn.Delete(path, DEFAULT_VERSION)
}

func (zm *ZkManager) RemoveInRecursive(path string) error {
	return recursiveDelete(zm.conn, path, true)
}

func (zm *ZkManager) UnWatch(path string) error {
	return nil
}
