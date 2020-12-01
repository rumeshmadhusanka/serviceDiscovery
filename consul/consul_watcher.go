package serviceDiscovery

import (
	"context"
	"errors"
	"time"
)

var (
	RequestTimeout = 5 * time.Second
	RepeatInterval = 5 * time.Second
)

type consulWatcher struct {
	client *consulClient
}

type Watcher interface {
	Watch(query *Query, doneChan <-chan int) (<-chan *NodeInfo, <-chan error)
}

//poll for a given query
func (c consulWatcher) Watch(query *Query, doneChan <-chan int) (<-chan *NodeInfo, <-chan error) {
	nodeInfoChan := make(chan *NodeInfo)
	errorChan := make(chan error)

	//start go routines
	//start them every 5 secs
	//control done, closing

	//long lived go routine
	go func() {
		ticker := time.NewTicker(RepeatInterval)
		intervalChan := ticker.C

		//cleanup if this go routine exits
		defer ticker.Stop()
		defer close(nodeInfoChan)
		defer close(errorChan)

		for {
			//todo find, do we really need a timout?
			timeout := time.After(RequestTimeout)
			ctx, cancel := context.WithCancel(context.Background())
			select {
			case <-intervalChan:
				go c.client.Nodes(ctx, query, nodeInfoChan, errorChan)
			case <-timeout:
				cancel()
				//todo add where the timout happened
				errorChan <- errors.New("timeout")
			case <-doneChan:
				cancel()
				return
			}
		}
	}()

	return nodeInfoChan, errorChan
}

//constructor
func NewConsulWatcher(client *consulClient) (*consulWatcher, error) {
	return &consulWatcher{client: client}, nil
}
