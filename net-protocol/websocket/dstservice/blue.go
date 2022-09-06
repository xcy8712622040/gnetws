/******************************
 * @Developer: many
 * @File: blue.go
 * @Time: 2022/6/6 14:40
******************************/

package dstservice

import (
	"fmt"
	"github.com/xcy8712622040/gnetws"
	"github.com/xcy8712622040/gnetws/eventserve"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket"
	"io"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
)

type Packet interface {
	io.Reader
	Head() string
}

type blueprint struct {
	gnetws.Serialize

	args         sync.Pool
	functionpool map[string]sync.Pool
}

func (self *blueprint) DoProc(ctx *eventserve.GnetContext, conn *websocket.Conn) {
	x := self.args.Get()
	defer self.args.Put(x)

	if err := self.NewDeCodec(conn.FrameReader()).Decode(x); err != nil {
		ctx.Logger.Errorf("DoProc Decode Error: %s", err)
	} else {
		conn.AwaitAdd()
		if err = gnetws.GoroutinePool().Submit(func() {
			defer func() {
				conn.Free()
				var exception interface{}
				switch exception = recover(); exception.(type) {
				case nil:
					return
				case runtime.Error:
					ctx.Logger.Errorf(string(debug.Stack()))
				default:
					ctx.Logger.Warnf("Do not use 'panic' in programs")
				}
				ctx.Logger.Errorf("Proc Call Error: %s", exception)
			}()

			pkt := x.(Packet)
			if f, ok := self.functionpool[pkt.Head()]; ok {
				data := f.Get()
				defer f.Put(data)
				if err = self.NewDeCodec(pkt).Decode(data); err == nil {
					handler := data.(websocket.Handler)
					text := conn.WebSocketTextWriter()
					if err = self.NewEnCodec(text).Encode(handler.Proc(ctx)); err == nil {
						if err = text.Flush(); err != nil {
							ctx.Logger.Errorf("WebSocketTextWriter Flush Error: %s", err)
						}
					} else {
						ctx.Logger.Errorf("Encode Error: %s", err)
					}
				} else {
					ctx.Logger.Errorf("Proc Decode Error: %s", err)
				}
			} else {
				ctx.Logger.Errorf("function [%s] non-existent", pkt.Head())
			}
		}); err != nil {
			conn.Free()
			ctx.Logger.Errorf("GoroutinePool Submit Error: %s", err)
		}
	}
}

func (self *blueprint) Route(head string, proc interface{}) (err error) {
	if _, ok := self.functionpool[head]; ok {
		err = fmt.Errorf("please note that, head [%s] has been overwritten", head)
	}

	refType := reflect.TypeOf(proc)
	self.functionpool[head] = sync.Pool{New: func() interface{} {
		return reflect.New(refType.Elem()).Interface()
	}}

	return
}
