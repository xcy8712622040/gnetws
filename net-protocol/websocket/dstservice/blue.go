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
	Proc
	Route(head string, proc interface{}) (err error)
}

type Packet interface {
	io.Reader
	Head() string
}

type blueprint struct {
	gnetws.Serialize

	argsPool     sync.Pool
	functionPool map[string]sync.Pool
}

func (b *blueprint) Args() (x interface{}) {
	return b.argsPool.Get()
}

func (b *blueprint) DoProc(ctx *serverhandler.Context, conn *websocket.Conn, x interface{}) {
	conn.AddCite()
	if err := gnetws.GoroutinePool().Submit(func() {
		defer func() {
			conn.Free()
			b.argsPool.Put(x)
		}()

		pkt := x.(Packet)
		if f, ok := b.functionPool[pkt.Head()]; ok {
			data := f.Get()
			defer f.Put(data)
			if err := b.NewDeCodec(pkt).Decode(data); err == nil {
				handler := data.(websocket.Handler)
				text := conn.WebSocketTextWriter()
				if err = b.NewEnCodec(text).Encode(call(ctx, handler)); err == nil {
					if err = text.Flush(); err != nil {
						ctx.Logger().Errorf("WebSocketTextWriter Flush Error: %s", err)
					}
				} else {
					ctx.Logger().Errorf("Encode Error: %s", err)
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
