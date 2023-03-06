package dstservice

import (
	"github.com/xcy8712622040/gnetws/serverhandler"
	"sync"
)

type (
	ServicePlugin interface{}

	// OnCallPrePlugin 逻辑处理之前
	OnCallPrePlugin interface {
		OnCallPre(ctx *serverhandler.Context, x interface{}) error
	}

	// OnCallPostPlugin 逻辑处理之后
	OnCallPostPlugin interface {
		OnCallPost(ctx *serverhandler.Context, x interface{}, reply interface{}) error
	}
)

type Plugins struct {
	mutex   sync.Mutex
	storage []ServicePlugin
}

func NewPlugins() *Plugins {
	return &Plugins{storage: []ServicePlugin{}}
}

func (p *Plugins) Add(plugin ServicePlugin) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.storage = append(p.storage, plugin)
}

func (p *Plugins) WsOnCallPre(ctx *serverhandler.Context, x interface{}) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for idx := range p.storage {
		if plug, ok := p.storage[idx].(OnCallPrePlugin); ok {
			if e := plug.OnCallPre(ctx, x); e != nil {
				return e
			}
		}
	}

	return nil
}

func (p *Plugins) WsOnCallPost(ctx *serverhandler.Context, x interface{}, reply interface{}) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for idx := range p.storage {
		if plug, ok := p.storage[idx].(OnCallPostPlugin); ok {
			if e := plug.OnCallPost(ctx, x, reply); e != nil {
				return e
			}
		}
	}

	return nil
}
