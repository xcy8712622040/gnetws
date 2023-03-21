package dstservice

import (
	"fmt"
	"github.com/xcy8712622040/gnetws"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket"
	"github.com/xcy8712622040/gnetws/serverhandler"
	"io"
	"reflect"
	"sync"
)

type Blueprint interface {
	Router

	Route(head string, proc interface{}) (err error)
}

type Packet interface {
	io.Reader
	Head() string
}

type blueprint struct {
	args         sync.Pool
	codex        gnetws.Serialize
	functionPool map[string]sync.Pool

	blueprintFunctionPlugins *Plugins
}

func (b *blueprint) Plugins() *Plugins {
	return b.blueprintFunctionPlugins
}

func (b *blueprint) Args() (x interface{}) {
	return b.args.Get()
}

func (b *blueprint) Codex() gnetws.Serialize {
	return b.codex
}

func (b *blueprint) DoProc(ctx *serverhandler.Context, conn *websocket.Conn, x interface{}) {
	conn.AddCite()
	if err := gnetws.GoroutinePool().Submit(func() {
		defer func() {
			conn.Free()
			b.args.Put(x)
		}()

		pkt := x.(Packet)
		if f, ok := b.functionPool[pkt.Head()]; ok {
			data := f.Get()
			defer f.Put(data)
			if err := b.Codex().NewDeCodec(pkt).Decode(data); err == nil {
				if err = b.Plugins().WsOnCallPre(ctx, data); err != nil {
					ctx.Logger().Errorf("ServicePlugin WsOnCallPre Error: %s", err)
				} else {
					reply := call(ctx, data.(websocket.Handler))
					if err = b.Plugins().WsOnCallPost(ctx, x, reply); err != nil {
						ctx.Logger().Errorf("ServicePlugin WsOnCallPost Error: %s", err)
					} else {
						text := conn.WebSocketTextWriter()
						if err = b.Codex().NewEnCodec(text).Encode(reply); err == nil {
							if err = text.Flush(); err != nil {
								ctx.Logger().Errorf("WebSocketTextWriter Flush Error: %s", err)
							}
						} else {
							ctx.Logger().Errorf("Encode Error: %s", err)
						}
					}
				}
			} else {
				ctx.Logger().Errorf("Proc Decode Error: %s", err)
			}
		} else {
			ctx.Logger().Errorf("function [%s] non-existent", pkt.Head())
		}
	}); err != nil {
		conn.Free()
		ctx.Logger().Errorf("GoroutinePool Submit Error: %s", err)
	}
}

func (b *blueprint) Route(head string, proc interface{}) (err error) {
	if _, ok := b.functionPool[head]; ok {
		err = fmt.Errorf("please note that, head [%s] has been overwritten", head)
	}

	refType := reflect.TypeOf(proc)
	b.functionPool[head] = sync.Pool{New: func() interface{} {
		return reflect.New(refType.Elem()).Interface()
	}}

	return
}
