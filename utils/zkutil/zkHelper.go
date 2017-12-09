package zkutil

import (
	"fmt"
	"strings"
	"time"

	"github.com/curator-go/curator"
	"github.com/samuel/go-zookeeper/zk"
)

func getRetryPolicy(maxRetryNum int, maxSleepTime time.Duration) curator.RetryPolicy {
	return curator.NewExponentialBackoffRetry(time.Millisecond, maxRetryNum, maxSleepTime)
}

func onConnectionStateChanged(f curator.CuratorFramework, e curator.ConnectionState) {
	switch e {
	case curator.CONNECTED:
		fmt.Println(e.String())
	case curator.RECONNECTED:
		fmt.Println(e.String())
	case curator.SUSPENDED:
		fmt.Println(e.String())
	case curator.LOST:
		fmt.Println(e.String())
	}
}

func connect(zm *ZkManager, maxRetryNum int, maxSleepTime, connectionTimeout, sessionTimeout time.Duration) error {
	retryPolicy := getRetryPolicy(maxRetryNum, maxSleepTime)
	zm.zkClient = newZkClientWithOptions(strings.Join(zm.MetaData.ZkAddr, ","), retryPolicy, connectionTimeout, sessionTimeout)
	err := zm.zkClient.Start()
	if err != nil {
		return err
	}
	return nil
}

func close(zm *ZkManager) error {
	return zm.zkClient.Close()
}

func addListeners(zm *ZkManager) {
	connListener := curator.NewConnectionStateListener(onConnectionStateChanged)
	listener := curator.NewCuratorListener(func(c curator.CuratorFramework, e curator.CuratorEvent) error {
		fmt.Println("listener type:", e.Type().String())
		// EventNodeCreated:         "EventNodeCreated",
		// EventNodeDeleted:         "EventNodeDeleted",
		// EventNodeDataChanged:     "EventNodeDataChanged",
		// EventNodeChildrenChanged: "EventNodeChildrenChanged",
		// EventSession:             "EventSession",
		// EventNotWatching:         "EventNotWatching",
		switch e.WatchedEvent().Type {
		case zk.EventNodeCreated:
			fmt.Println("watchevent:", e.WatchedEvent())
		case zk.EventNodeDeleted:
			fmt.Println("watchevent:", e.WatchedEvent())
		case zk.EventNodeDataChanged:
			fmt.Println("watchevent:", e.WatchedEvent(), e.Data())
			err := zm.GetNodeDataW(e.Path(), onEventNodeDataChanged)
			if err != nil {
				// todo
			}
		case zk.EventNodeChildrenChanged:
			fmt.Println("watchevent:", e.WatchedEvent(), e.Children(), e.Data())
			err := zm.GetChildrenW(e.Path(), onEventNodeChildrenChanged)
			if err != nil {
				// todo
			}
		}

		return nil
	})

	zm.AddConnectionListener(connListener)
	zm.AddListener(listener)
}

func newZkClient(connString string, maxRetryNum int, maxSleepTime time.Duration) curator.CuratorFramework {
	// these are reasonable arguments for the ExponentialBackoffRetry.
	// the first retry will wait 1 second,
	// the second will wait up to 2 seconds,
	// the third will wait up to 4 seconds.
	retryPolicy := curator.NewExponentialBackoffRetry(time.Millisecond, maxRetryNum, maxSleepTime)

	// The simplest way to get a CuratorFramework instance. This will use default values.
	// The only required arguments are the connection string and the retry policy
	return curator.NewClient(connString, retryPolicy)
}

func newZkClientWithOptions(connString string, retryPolicy curator.RetryPolicy, connectionTimeout, sessionTimeout time.Duration) curator.CuratorFramework {
	// using the CuratorFrameworkBuilder gives fine grained control over creation options.
	builder := &curator.CuratorFrameworkBuilder{
		ConnectionTimeout: connectionTimeout,
		SessionTimeout:    sessionTimeout,
		RetryPolicy:       retryPolicy,
	}

	// return builder.ConnectString(connString).Authorization("digest", []byte("user:pass")).Build()
	return builder.ConnectString(connString).Build()
}
