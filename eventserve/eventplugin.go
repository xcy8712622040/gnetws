package eventserve

import (
	"bytes"
	"github.com/panjf2000/gnet/v2"
	"sync"
)

type EventPlugin interface{}

type (
	// 服务器启动之后
	OnBootPlugin interface {
		OnBoot(eng gnet.Engine) (action gnet.Action)
	}

	// 服务器关闭之前
	OnShutdownPlugin interface {
		OnShutdown(eng gnet.Engine)
	}

	// 链接建立之后
	OnOpenPlugin interface {
		OnOpen(c *GnetContext) (out []byte, action gnet.Action)
	}

	// 链接关闭之前
	OnClosePlugin interface {
		OnClose(c *GnetContext, err error) (action gnet.Action)
	}

	// 数据到来
	OnTrafficPrePlugin interface {
		OnTrafficPre(c *GnetContext, size int) (action gnet.Action)
	}
)

type plugins struct {
	mutex   sync.Mutex
	storage []EventPlugin
}

func (self *plugins) Add(plugin EventPlugin) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.storage = append(self.storage, plugin)
}

func (self *plugins) EventServerOnBoot(eng gnet.Engine) (action gnet.Action) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	for idx := range self.storage {
		plugin := self.storage[idx]
		if bkplugin, ok := plugin.(OnBootPlugin); ok {
			if e := bkplugin.OnBoot(eng); e != gnet.None {
				return e
			}
		}
	}

	return gnet.None
}

func (self *plugins) EventServerOnShutdown(eng gnet.Engine) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	for idx := range self.storage {
		plugin := self.storage[idx]
		if bkplugin, ok := plugin.(OnShutdownPlugin); ok {
			bkplugin.OnShutdown(eng)
		}
	}
}

func (self *plugins) EventServerOnOpen(c *GnetContext) (out []byte, action gnet.Action) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	w := bytes.NewBuffer(out)
	for idx := range self.storage {
		plugin := self.storage[idx]
		if bkplugin, ok := plugin.(OnOpenPlugin); ok {
			if buf, act := bkplugin.OnOpen(c); act != gnet.None {
				w.Write(buf)
				return out[:w.Len()], act
			} else {
				w.Write(buf)
			}
		}
	}
	return out[:w.Len()], action
}

func (self *plugins) EventServerOnClose(c *GnetContext, e error) (action gnet.Action) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	for idx := range self.storage {
		plugin := self.storage[idx]
		if bkplugin, ok := plugin.(OnClosePlugin); ok {
			if act := bkplugin.OnClose(c, e); act != gnet.None {
				return act
			}
		}
	}
	return action
}

func (self *plugins) EventServerOnTrafficPre(c *GnetContext, size int) (action gnet.Action) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	for idx := range self.storage {
		plugin := self.storage[idx]
		if bkplugin, ok := plugin.(OnTrafficPrePlugin); ok {
			if act := bkplugin.OnTrafficPre(c, size); act != gnet.None {
				return act
			}
		}
	}
	return action
}
