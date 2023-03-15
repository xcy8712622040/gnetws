package serverhandler

import (
	"bytes"
	"github.com/panjf2000/gnet/v2"
	"sync"
)

type EventPlugin interface{}

type (
	// OnBootPlugin 服务器启动之后
	OnBootPlugin interface {
		OnBoot(eng gnet.Engine) (action gnet.Action)
	}

	// OnShutdownPlugin 服务器关闭之前
	OnShutdownPlugin interface {
		OnShutdown(eng gnet.Engine)
	}

	// OnOpenPlugin 链接建立之后
	OnOpenPlugin interface {
		OnOpen(c *Context) (out []byte, action gnet.Action)
	}

	// OnClosePlugin 链接关闭之前
	OnClosePlugin interface {
		OnClose(c *Context, err error) (action gnet.Action)
	}

	// OnTrafficPrePlugin 数据到来
	OnTrafficPrePlugin interface {
		OnTrafficPre(c *Context, size int) (action gnet.Action)
	}
)

type plugins struct {
	mutex   sync.Mutex
	storage []EventPlugin
}

func (p *plugins) Add(plugin EventPlugin) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.storage = append(p.storage, plugin)
}

func (p *plugins) EventServerOnBoot(eng gnet.Engine) (action gnet.Action) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for idx := range p.storage {
		plugin := p.storage[idx]
		if bootPlugin, ok := plugin.(OnBootPlugin); ok {
			if e := bootPlugin.OnBoot(eng); e != gnet.None {
				return e
			}
		}
	}

	return gnet.None
}

func (p *plugins) EventServerOnShutdown(eng gnet.Engine) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for idx := range p.storage {
		plugin := p.storage[idx]
		if shutdownPlugin, ok := plugin.(OnShutdownPlugin); ok {
			shutdownPlugin.OnShutdown(eng)
		}
	}
}

func (p *plugins) EventServerOnOpen(c *Context) (out []byte, action gnet.Action) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	w := bytes.NewBuffer(out)
	for idx := range p.storage {
		plugin := p.storage[idx]
		if openPlugin, ok := plugin.(OnOpenPlugin); ok {
			if buf, act := openPlugin.OnOpen(c); act != gnet.None {
				w.Write(buf)
				return out[:w.Len()], act
			} else {
				w.Write(buf)
			}
		}
	}
	return out[:w.Len()], action
}

func (p *plugins) EventServerOnClose(c *Context, e error) (action gnet.Action) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for idx := range p.storage {
		plugin := p.storage[idx]
		if closePlugin, ok := plugin.(OnClosePlugin); ok {
			if act := closePlugin.OnClose(c, e); act != gnet.None {
				return act
			}
		}
	}
	return action
}

func (p *plugins) EventServerOnTrafficPre(c *Context, size int) (action gnet.Action) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for idx := range p.storage {
		plugin := p.storage[idx]
		if trafficPre, ok := plugin.(OnTrafficPrePlugin); ok {
			if act := trafficPre.OnTrafficPre(c, size); act != gnet.None {
				return act
			}
		}
	}
	return action
}
