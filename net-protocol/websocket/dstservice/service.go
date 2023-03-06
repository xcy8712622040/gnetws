package dstservice

import (
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/panjf2000/gnet/v2"
	"github.com/xcy8712622040/gnetws"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket"
	"github.com/xcy8712622040/gnetws/serverhandler"
	"io"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
)

const (
	RequestHeader = "__request_header__"
)

type Proc interface {
	gnetws.Serialize

	Plugins() *Plugins

	Args() (x interface{})
	DoProc(ctx *serverhandler.Context, conn *websocket.Conn, x interface{})
}

func call(ctx *serverhandler.Context, hand websocket.Handler) interface{} {
	defer func() {
		var exception interface{}
		switch exception = recover(); exception.(type) {
		case nil:
			return
		case runtime.Error:
			ctx.Logger().Errorf(string(debug.Stack()))
		default:
			ctx.Logger().Warnf("Do not use 'panic' in programs")
		}
		ctx.Logger().Errorf("Proc Call Error: %s", exception)
	}()
	return hand.Proc(ctx)
}

type root struct {
	gnetws.Serialize
	argsPool sync.Pool

	plugins *Plugins
}

func (r *root) Plugins() *Plugins { return r.plugins }

func (r *root) Args() (x interface{}) {
	return r.argsPool.Get()
}

func (r *root) DoProc(ctx *serverhandler.Context, conn *websocket.Conn, x interface{}) {
	conn.AddCite()
	if err := gnetws.GoroutinePool().Submit(func() {
		defer func() {
			conn.Free()
			r.argsPool.Put(x)
		}()

		if err := r.plugins.WsOnCallPre(ctx, x); err != nil {
			ctx.Logger().Errorf("ServicePlugin WsOnCallPre Error: %s", err)
		} else {
			reply := call(ctx, x.(websocket.Handler))
			if err = r.plugins.WsOnCallPost(ctx, x, reply); err != nil {
				ctx.Logger().Errorf("ServicePlugin WsOnCallPost Error: %s", err)
			} else {
				text := conn.WebSocketTextWriter()
				if err = r.NewEnCodec(text).Encode(reply); err == nil {
					if err = text.Flush(); err != nil {
						ctx.Logger().Errorf("WebSocketTextWriter Flush Error: %s", err)
					}
				} else {
					ctx.Logger().Errorf("Encode Error: %s", err)
				}
			}
		}
	}); err != nil {
		conn.Free()
		ctx.Logger().Errorf("GoroutinePool Submit Error: %s", err)
	}
}

type WebSocketFrameHandler struct {
	method Proc

	Plugins *websocket.Plugins
}

func (w *WebSocketFrameHandler) WithConn(ctx *serverhandler.Context, conn gnet.Conn) error {
	frame := websocket.FrameConvert(conn)

	defer frame.Free()
	for err := frame.NextFrame(); err != io.EOF; err = frame.NextFrame() {
		if err != nil {
			return err
		}

		x := w.method.Args()
		if err = w.method.NewDeCodec(frame.FrameReader()).Decode(x); err == nil {
			w.method.DoProc(ctx, frame, x)
		} else {
			ctx.Logger().Errorf("DoProc Error: %s", err)
		}
	}

	return nil
}

type WebSocketUpgradeHandle struct {
	WebSocketFrameHandler

	methodMap map[string]Proc
}

func (w *WebSocketUpgradeHandle) Blueprint(path string, args Packet, codec gnetws.Serialize) Blueprint {
	refType := reflect.TypeOf(args)
	w.methodMap[path] = &blueprint{
		Serialize:    codec,
		plugins:      NewPlugins(),
		functionPool: map[string]sync.Pool{}, argsPool: sync.Pool{
			New: func() interface{} {
				return reflect.New(refType.Elem()).Interface()
			},
		}}
	return w.methodMap[path].(*blueprint)
}

func (w *WebSocketUpgradeHandle) Route(path string, codec gnetws.Serialize, proc websocket.Handler) (route Proc, err error) {
	if _, ok := w.methodMap[path]; ok {
		err = fmt.Errorf("please note that, path [%s] has been overwritten", path)
	}

	refType := reflect.TypeOf(proc)
	w.methodMap[path] = &root{
		Serialize: codec,
		plugins:   NewPlugins(),
		argsPool: sync.Pool{New: func() interface{} {
			return reflect.New(refType.Elem()).Interface()
		}},
	}

	return w.methodMap[path], err
}

func (w *WebSocketUpgradeHandle) WithConn(ctx *serverhandler.Context, conn gnet.Conn) error {
	up := ws.Upgrader{}

	header := map[string]string{}

	wsFrameHandler := w.WebSocketFrameHandler

	up.OnRequest = func(uri []byte) (err error) {
		if method, ok := w.methodMap[string(uri)]; ok {
			wsFrameHandler.method = method
		} else {
			err = errors.New(fmt.Sprintf("uri [%s] no counterpart handler", string(uri)))
			ctx.Logger().Errorf("ws upgrade error: %s", err)
		}
		return
	}

	up.OnHeader = func(key, value []byte) error {
		header[string(key)] = string(value)
		return nil
	}

	if _, err := up.Upgrade(conn); err != nil {
		return err
	} else {
		ctx.WithHandler(&wsFrameHandler)
	}
	ctx.MetaData().SetInterface(RequestHeader, header)

	return w.Plugins.WsOnUpgrade(ctx, conn)
}

var GlobalService = NewWebSocketHandler()

func NewWebSocketHandler() *WebSocketUpgradeHandle {
	return &WebSocketUpgradeHandle{
		methodMap: map[string]Proc{},
		WebSocketFrameHandler: WebSocketFrameHandler{
			Plugins: websocket.NewPlugins(),
		},
	}
}
