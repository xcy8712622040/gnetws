package websocket

import (
	"github.com/panjf2000/gnet/v2"
	"github.com/xcy8712622040/gnetws/serverhandler"
	"sync"
)

type (
	WsPlugin interface{}

	// OnUpgradePlugin ws链接建立成功之后
	OnUpgradePlugin interface {
		OnUpgrade(ctx *serverhandler.Context, conn gnet.Conn, path string) error
	}

	// OnClosedPlugin ws帧解析错误 关闭 ws链接之前
	OnClosedPlugin interface {
		OnWsClosed(ctx *serverhandler.Context, err error)
	}
)

type Plugins struct {
	mutex   sync.Mutex
	storage []WsPlugin
}

func NewPlugins() *Plugins {
	return &Plugins{storage: []WsPlugin{}}
}

func (p *Plugins) Add(plugin WsPlugin) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.storage = append(p.storage, plugin)
}

func (p *Plugins) WsOnUpgrade(ctx *serverhandler.Context, conn gnet.Conn, path string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for idx := range p.storage {
		if plug, ok := p.storage[idx].(OnUpgradePlugin); ok {
			if e := plug.OnUpgrade(ctx, conn, path); e != nil {
				return e
			}
		}
	}

	return nil
}

func (p *Plugins) WsOnClosedPlugin(ctx *serverhandler.Context, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for idx := range p.storage {
		if plug, ok := p.storage[idx].(OnClosedPlugin); ok {
			plug.OnWsClosed(ctx, err)
		}
	}
}
