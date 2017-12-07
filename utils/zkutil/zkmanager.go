package zkutil

import (
	"finder-go/common"
	"finder-go/companion"
	"finder-go/utils/arrayutil"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/curator-go/curator"
	"github.com/samuel/go-zookeeper/zk"
)

var (
	hc               *http.Client
	url              string
	zkExit           chan bool
	ConfigEventPool  *ConfigChangedEventPool
	ServiceEventPool *ServiceChangedEventPool
)

func init() {
	hc = &http.Client{
		Transport: &http.Transport{
			Dial: func(nw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(1 * time.Second)
				c, err := net.DialTimeout(nw, addr, time.Second*1)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	ConfigEventPool = NewConfigChangedEventPool()
	ServiceEventPool = NewServiceChangedEventPool()
}

// ZkManager for operate zk
type ZkManager struct {
	MetaData          *common.ZkInfo
	checkZkInfoTicker *time.Ticker
	zkClient          curator.CuratorFramework
}

// NewZkManager for create ZkManager
func NewZkManager(config *common.BootConfig) (*ZkManager, error) {
	checkConfig(config)
	zm, err := init_(config)
	if err != nil {
		return nil, err
	}
	zm.checkZkInfoTicker = time.NewTicker(config.TickerDuration)
	// 开启一个协程去检测zkinfo变化
	go watchZkAddr(zm)
	// 创建zk连接
	err = connect(zm, config.ZkMaxRetryNum, config.ZkMaxSleepTime, config.ZkConnectTimeout, config.ZkSessionTimeout)
	if err != nil {
		return nil, err
	}
	// 增加监听
	addListeners(zm)

	return zm, nil
}

func checkConfig(c *common.BootConfig) {
	if c.TickerDuration <= 0 {
		c.ZkConnectTimeout = 30 * time.Second
	}
	if c.ZkConnectTimeout <= 0 {
		c.ZkConnectTimeout = 3 * time.Second
	}
	if c.ZkSessionTimeout <= 0 {
		c.ZkSessionTimeout = 30 * time.Second
	}
	if c.ZkMaxRetryNum < 0 {
		c.ZkMaxRetryNum = 3
	}
	if c.ZkMaxSleepTime <= 0 {
		c.ZkMaxSleepTime = 15 * time.Second
	}
}

func init_(config *common.BootConfig) (*ZkManager, error) {
	url = config.CompanionUrl + fmt.Sprintf("/finder/query_zk_info?project=%s&group=%s&service=%s&version=%s", config.MeteData.Project, config.MeteData.Group, config.MeteData.Service, config.MeteData.Version)
	metadata, err := companion.GetZkInfo(hc, url)
	if err != nil {
		return nil, err
	}
	zm := &ZkManager{
		MetaData: metadata,
	}

	return zm, nil
}

func (zm *ZkManager) CreatePath(path string) (string, error) {
	return zm.zkClient.Create().CreatingParentsIfNeeded().ForPath(path)
}

func (zm *ZkManager) CreatePathWithData(path string, data []byte) (string, error) {
	return zm.zkClient.Create().CreatingParentsIfNeeded().ForPathWithData(path, data)
}

func (zm *ZkManager) CreateTempPath(path string) (string, error) {
	return zm.zkClient.Create().CreatingParentsIfNeeded().WithMode(curator.EPHEMERAL).ForPath(path)
}

func (zm *ZkManager) CreateTempPathWithData(path string, data []byte) (string, error) {
	return zm.zkClient.Create().CreatingParentsIfNeeded().WithMode(curator.EPHEMERAL).ForPathWithData(path, data)
}

func (zm *ZkManager) UpdateData(path string, data []byte) (*zk.Stat, error) {
	return zm.zkClient.SetData().Compressed().ForPathWithData(path, data)
}

func (zm *ZkManager) ExistsNode(path string) (*zk.Stat, error) {
	return zm.zkClient.CheckExists().ForPath(path)
}

func (zm *ZkManager) ExistsNodeW(path string) (*zk.Stat, error) {
	return zm.zkClient.CheckExists().Watched().ForPath(path)
}

func (zm *ZkManager) UpdateDataWithCheckExists(path string, data []byte) (*zk.Stat, error) {
	s, err := zm.zkClient.CheckExists().ForPath(path)
	if err != nil {
		return nil, err
	}
	if s != nil {
		return zm.zkClient.SetData().Compressed().ForPathWithData(path, data)
	}

	return s, nil
}

func (zm *ZkManager) GetNodeData(path string) ([]byte, error) {
	// return zm.zkClient.GetData().Decompressed().ForPath(path)
	return zm.zkClient.GetData().ForPath(path)
}

func (zm *ZkManager) GetNodeDataW(path string, c curator.BackgroundCallback) error {
	// return zm.zkClient.GetData().Decompressed().UsingWatcher(watcher).ForPath(path)
	fmt.Println(path)
	_, err := zm.zkClient.GetData().InBackgroundWithCallback(c).Watched().ForPath(path)
	return err
}

func (zm *ZkManager) GetChildrenNodeUseWatch(path string, watcher curator.Watcher) ([]string, error) {
	return zm.zkClient.GetChildren().UsingWatcher(watcher).ForPath(path)
}

func (zm *ZkManager) GetChildrenW(path string, c curator.BackgroundCallback) error {
	_, err := zm.zkClient.GetChildren().InBackgroundWithCallback(c).Watched().ForPath(path)
	return err
}

func (zm *ZkManager) GetChildren(path string) error {
	_, err := zm.zkClient.GetChildren().ForPath(path)
	return err
}

func (zm *ZkManager) RemoveInRecursive(path string) error {
	return zm.zkClient.Delete().DeletingChildrenIfNeeded().ForPath(path)
}

func (zm *ZkManager) AddListener(listener curator.CuratorListener) {
	zm.zkClient.CuratorListenable().AddListener(listener)
}

func (zm *ZkManager) AddConnectionListener(listener curator.ConnectionStateListener) {
	zm.zkClient.ConnectionStateListenable().AddListener(listener)
}

func (zm *ZkManager) RemoveListener(listener curator.CuratorListener) {
	zm.zkClient.CuratorListenable().RemoveListener(listener)
}

func (zm *ZkManager) Destroy() {
	zkExit <- true
	zm.checkZkInfoTicker.Stop()
	err := close(zm)
	if err != nil {

	}
}

func onZkInfoChanged(zm *ZkManager) {
	// todo.
}

func onEventNodeChildrenChanged(c curator.CuratorFramework, e curator.CuratorEvent) error {
	event, ok := ServiceEventPool.Get()[e.Name()]
	if ok {
		s := common.Service{
			Name:       e.Name(),
			ServerList: getServiceItems(c, e.Path(), e.Children()),
		}

		event(s)
	}

	return nil
}

func onEventNodeCreated(e *zk.Event) {

}

func onEventNodeDataChanged(c curator.CuratorFramework, e curator.CuratorEvent) error {
	event, ok := ConfigEventPool.Get()[e.Name()]
	if ok {
		pushId, fData, err := DecodeValue(e.Data())
		if err != nil {
			// todo
		} else {
			c := common.Config{
				PushId: pushId,
				Name:   e.Name(),
				File:   fData,
			}

			ok := event(c)
			if ok {

			}
			// todo feedback
		}
	}

	return nil
}

func onEventNodeDeleted(e *zk.Event) {

}

func onEventNotWatching(e *zk.Event) {

}

func getServiceItems(c curator.CuratorFramework, parentPath string, children []string) []common.ServiceItem {
	serverList := make([]common.ServiceItem, 0)
	// var data []byte
	// var err error
	for _, n := range children {
		// data, err = c.GetData().ForPath(parentPath + "/" + n)
		// if err != nil {
		// 	continue
		// }
		serverList = append(serverList, common.ServiceItem{Addr: n, Weight: 100, IsValid: true})
	}

	return serverList
}

func watchZkAddr(zm *ZkManager) {
	for t := range zm.checkZkInfoTicker.C {
		//fmt.Println(t)
		if t.IsZero() {

		}
		metadata, err := companion.GetZkInfo(hc, url)
		if err != nil {
			// todo.
			continue
		}
		vchanged := checkAddr(metadata.ZkAddr, zm.MetaData.ZkAddr)
		if vchanged {
			zm.MetaData.ZkAddr = metadata.ZkAddr
			zm.MetaData.ConfigRootPath = metadata.ConfigRootPath
			zm.MetaData.ServiceRootPath = metadata.ServiceRootPath
			// 通知zkinfo更新，执行相关逻辑
			onZkInfoChanged(zm)
		}
	}
}

func checkAddr(n []string, o []string) bool {
	vchanged := false
	for _, nv := range o {
		if !arrayutil.Contains(nv, o) {
			vchanged = true
		}
	}

	return vchanged
}