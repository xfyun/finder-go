package zookeeper

import (
	"github.com/cooleric/go-zookeeper/zk"
)

func eventCallback(e zk.Event) {
	//log.Println(e.Type, "_", e.State, "_", e.Server, "_", e.Path)

	switch e.Type {
	case zk.EventNodeCreated:
		return
	case zk.EventNodeDeleted:
		return
	case zk.EventNodeDataChanged:
		return
	case zk.EventNodeChildrenChanged:
		return
	case zk.EventNotWatching:
		return
	case zk.EventSession:
		catchSessionState(e)
		return
	}
}

func catchSessionState(e zk.Event) {
	switch e.State {
	case zk.StateDisconnected:
		return
	case zk.StateConnecting:
		return
	case zk.StateConnected:
		return
	case zk.StateHasSession:
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
}
