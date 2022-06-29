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

type function struct {
	args *sync.Pool
	proc func(ctx *eventserve.WebSocketContext, args interface{}) interface{}
}

type blueprint struct {
	DoCodec

	args         *sync.Pool
	functionpool map[string]function
}

func (self *blueprint) DoProc(ctx *eventserve.WebSocketContext, conn *eventserve.Conn) {
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
				data := f.args.Get()
				defer f.args.Put(data)
				if err = self.NewDeCodec(pkt).Decode(data); err == nil {

					text := conn.WebSocketTextWriter()
					if err = self.NewEnCodec(text).Encode(f.proc(ctx, x)); err == nil {
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

func (self *blueprint) Route(head string, args interface{}, proc func(ctx *eventserve.WebSocketContext, args interface{}) interface{}) (err error) {
	if _, ok := self.functionpool[head]; ok {
		err = fmt.Errorf("please note that, head [%s] has been overwritten", head)
	}

	refType := reflect.TypeOf(args)
	self.functionpool[head] = function{proc: proc, args: &sync.Pool{New: func() interface{} {
		if refType.Kind() != reflect.Ptr {
			return reflect.New(refType).Interface()
		} else {
			return reflect.New(refType.Elem()).Interface()
		}
	}}}

	return
}
