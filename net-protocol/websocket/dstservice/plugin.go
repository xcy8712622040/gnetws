package dstservice

import (
	"github.com/xcy8712622040/gnetws/eventserve"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket"
	"sync"
)

type plugins struct {
	mutex   sync.Mutex
	storage []websocket.WsPlugin
}

func (self *plugins) Add(plugin websocket.WsPlugin) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.storage = append(self.storage, plugin)
}

func (self *plugins) WsOnUpgrader(ctx *eventserve.GnetContext) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	for idx := range self.storage {
		if plug, ok := self.storage[idx].(websocket.OnUpgraderPlugin); ok {
			if e := plug.OnUpgrader(ctx); e != nil {
				return e
			}
		}
	}

	return nil
}

func (self *plugins) WsOnCallPre(ctx *eventserve.GnetContext, x interface{}) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	for idx := range self.storage {
		if plug, ok := self.storage[idx].(websocket.OnCallPrePlugin); ok {
			if e := plug.OnCallPre(ctx, x); e != nil {
				return e
			}
		}
	}

	return nil
}

func (self *plugins) WsOnCallPost(ctx *eventserve.GnetContext, x interface{}, reply interface{}) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	for idx := range self.storage {
		if plug, ok := self.storage[idx].(websocket.OnCallPostPlugin); ok {
			if e := plug.OnCallPost(ctx, x, reply); e != nil {
				return e
			}
		}
	}

	return nil
}
