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
	"github.com/xcy8712622040/gnetws/net-protocol/websocket"
	"io"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
)

const (
	RequestHeader = "__request_header__"
)

type Method interface {
	gnetws.Serialize
	DoProc(ctx *eventserve.GnetContext, conn *websocket.Conn)
}

type root struct {
	gnetws.Serialize
	handler sync.Pool
}

func (self *root) DoProc(ctx *eventserve.GnetContext, conn *websocket.Conn) {
	x := self.handler.Get()
	if err := self.NewDeCodec(conn.FrameReader()).Decode(x); err == nil {
		conn.AwaitAdd()
		if err = gnetws.GoroutinePool().Submit(func() {
			defer func() {
				conn.Free()
				self.handler.Put(x)
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

			handler := x.(websocket.Handler)
			text := conn.WebSocketTextWriter()
			if err = self.NewEnCodec(text).Encode(handler.Proc(ctx)); err == nil {
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

func (self *WebSocketHandler) Proc(ctx *eventserve.GnetContext, conn gnet.Conn) error {
	frame := websocket.FrameConvert(conn)

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

func (self *WebSocketHandler) Blueprint(path string, args Packet, codec gnetws.Serialize) *blueprint {
	refType := reflect.TypeOf(args)
	self.methodpool[path] = &blueprint{Serialize: codec, functionpool: map[string]sync.Pool{}, args: sync.Pool{
		New: func() interface{} {
			return reflect.New(refType.Elem()).Interface()
		},
	}}
	return self.methodpool[path].(*blueprint)
}

func (self *WebSocketHandler) Route(path string, codec gnetws.Serialize, proc websocket.Handler) (err error) {
	if _, ok := self.methodpool[path]; ok {
		err = fmt.Errorf("please note that, path [%s] has been overwritten", path)
	}

	refType := reflect.TypeOf(proc)
	self.methodpool[path] = &root{Serialize: codec, handler: sync.Pool{New: func() interface{} {
		return reflect.New(refType.Elem()).Interface()
	}}}

	return
}

type WebSocketUP struct {
	ws.Upgrader
	WebSocketHandler
}

func (self *WebSocketUP) Proc(ctx *eventserve.GnetContext, conn gnet.Conn) error {
	up := self.Upgrader
	wh := self.WebSocketHandler
	header := map[string]string{}
	matedata := ctx.Value(eventserve.MATEDATA).(*eventserve.MateData)

	up.OnRequest = func(uri []byte) error {
		wh.Path = string(uri)
		return nil
	}

	up.OnHeader = func(key, value []byte) error {
		header[string(key)] = string(value)
		return nil
	}

	if _, err := up.Upgrade(conn); err != nil {
		return err
	} else {
		ctx.DoProc = &wh
	}
	matedata.SetInterface(RequestHeader, header)

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
