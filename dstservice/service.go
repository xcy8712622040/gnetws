/******************************
 * @Developer: many
 * @File: dstservice.go
 * @Time: 2022/5/25 17:45
******************************/

package dstservice

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/panjf2000/gnet/v2"
	"github.com/xcy8712622040/gnetws"
	"github.com/xcy8712622040/gnetws/eventserve"
	"io"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
)

type Method interface {
	gnetws.DoCodec
	DoProc(ctx *eventserve.WebSocketContext, conn *eventserve.Conn)
}

type root struct {
	gnetws.DoCodec

	args *sync.Pool
	proc func(ctx *eventserve.WebSocketContext, args interface{}) interface{}
}

func (self *root) DoProc(ctx *eventserve.WebSocketContext, conn *eventserve.Conn) {
	x := self.args.Get()
	if err := self.NewDeCodec(conn.FrameReader()).Decode(x); err == nil {
		conn.AwaitAdd()
		if err = gnetws.GoroutinePool().Submit(func() {
			defer func() {
				conn.Free()
				self.args.Put(x)
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

			text := conn.WebSocketTextWriter()
			if err = self.NewEnCodec(text).Encode(self.proc(ctx, x)); err == nil {
				if err = text.Flush(); err != nil {
					ctx.Logger.Errorf("WebSocketTextWriter Flush Error: %s", err)
				}
			} else {
				ctx.Logger.Errorf("Encode Error: %s", err)
			}
		}); err != nil {
			conn.Free()
			ctx.Logger.Errorf("GoroutinePool Submit Error: %s", err)
		}
	} else {
		ctx.Logger.Errorf("DoProc Error: %s", err)
	}
}

type WebSocketHandler struct {
	Path       string
	methodpool map[string]Method
}

func (self *WebSocketHandler) Proc(ctx *eventserve.WebSocketContext, conn gnet.Conn) error {
	frame := eventserve.FrameConvert(conn)

	defer frame.Free()
	for err := frame.NextFrame(); err != io.EOF; err = frame.NextFrame() {
		if err != nil {
			return err
		}

		if method, ok := self.methodpool[self.Path]; ok {
			method.DoProc(ctx, frame)
		} else {
			return fmt.Errorf("path [ %s ] non-existent", self.Path)
		}
	}

	return nil
}

func (self *WebSocketHandler) Blueprint(path string, args Packet, codec gnetws.DoCodec) *blueprint {
	refType := reflect.TypeOf(args)
	self.methodpool[path] = &blueprint{DoCodec: codec, functionpool: map[string]function{}, args: &sync.Pool{
		New: func() interface{} {
			if refType.Kind() != reflect.Ptr {
				return reflect.New(refType).Interface()
			} else {
				return reflect.New(refType.Elem()).Interface()
			}
		},
	}}
	return self.methodpool[path].(*blueprint)
}

func (self *WebSocketHandler) Route(path string, args interface{}, codec gnetws.DoCodec, proc func(ctx *eventserve.WebSocketContext, args interface{}) interface{}) (err error) {
	if _, ok := self.methodpool[path]; ok {
		err = fmt.Errorf("please note that, path [%s] has been overwritten", path)
	}

	refType := reflect.TypeOf(args)
	self.methodpool[path] = &root{DoCodec: codec, proc: proc, args: &sync.Pool{New: func() interface{} {
		if refType.Kind() != reflect.Ptr {
			return reflect.New(refType).Interface()
		} else {
			return reflect.New(refType.Elem()).Interface()
		}
	}}}

	return
}

type WebSocketUP struct {
	ws.Upgrader
	WebSocketHandler
}

func (self *WebSocketUP) Proc(ctx *eventserve.WebSocketContext, conn gnet.Conn) error {
	up := self.Upgrader
	wh := self.WebSocketHandler

	up.OnRequest = func(uri []byte) error {
		wh.Path = string(uri)
		return nil
	}

	if _, err := up.Upgrade(conn); err != nil {
		return err
	} else {
		ctx.DoProc = &wh
	}

	return nil
}

func NewWebSocketHandler() *WebSocketUP {
	return &WebSocketUP{
		WebSocketHandler: WebSocketHandler{
			methodpool: map[string]Method{},
		},
	}
}

var Handler = NewWebSocketHandler()
