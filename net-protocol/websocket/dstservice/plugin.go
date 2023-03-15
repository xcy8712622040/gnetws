package dstservice

import (
	"github.com/xcy8712622040/gnetws/net-protocol/websocket"
	"github.com/xcy8712622040/gnetws/serverhandler"
	"sync"
)

type plugins struct {
	mutex   sync.Mutex
	storage []websocket.WsPlugin
}

func (p *plugins) Add(plugin websocket.WsPlugin) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.storage = append(p.storage, plugin)
}

func (p *plugins) WsOnUpgrade(ctx *serverhandler.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for idx := range p.storage {
		if plug, ok := p.storage[idx].(websocket.OnUpgradePlugin); ok {
			if e := plug.OnUpgrade(ctx); e != nil {
				return e
			}
		}
	}

	return nil
}

func (p *plugins) WsOnCallPre(ctx *serverhandler.Context, x interface{}) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for idx := range p.storage {
		if plug, ok := p.storage[idx].(websocket.OnCallPrePlugin); ok {
			if e := plug.OnCallPre(ctx, x); e != nil {
				return e
			}
		}
	}

	return nil
}

func (p *plugins) WsOnCallPost(ctx *serverhandler.Context, x interface{}, reply interface{}) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for idx := range p.storage {
		if plug, ok := p.storage[idx].(websocket.OnCallPostPlugin); ok {
			if e := plug.OnCallPost(ctx, x, reply); e != nil {
				return e
			}
		}
	}

	return nil
}
