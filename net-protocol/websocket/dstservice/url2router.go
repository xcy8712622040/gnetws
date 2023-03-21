package dstservice

import (
	"fmt"
	"github.com/xcy8712622040/gnetws"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket"
	"github.com/xcy8712622040/gnetws/serverhandler"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
)

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

type Router interface {
	websocket.Proc

	Plugins() *Plugins
}

type router struct {
	args  sync.Pool
	codex gnetws.Serialize

	routerMethodPlugins *Plugins
}

func (r *router) Plugins() *Plugins {
	return r.routerMethodPlugins
}

func (r *router) Args() (x interface{}) {
	return r.args.Get()
}

func (r *router) Codex() gnetws.Serialize {
	return r.codex
}

func (r *router) DoProc(ctx *serverhandler.Context, conn *websocket.Conn, x interface{}) {
	conn.AddCite()
	if err := gnetws.GoroutinePool().Submit(func() {
		defer func() {
			conn.Free()
			r.args.Put(x)
		}()

		if err := r.Plugins().WsOnCallPre(ctx, x); err != nil {
			ctx.Logger().Errorf("ServicePlugin WsOnCallPre Error: %s", err)
		} else {
			reply := call(ctx, x.(websocket.Handler))
			if err = r.Plugins().WsOnCallPost(ctx, x, reply); err != nil {
				ctx.Logger().Errorf("ServicePlugin WsOnCallPost Error: %s", err)
			} else {
				text := conn.WebSocketTextWriter()
				if err = r.Codex().NewEnCodec(text).Encode(reply); err == nil {
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

type Service struct {
	url2method map[string]Router
}

var GlobalService = NewUrlRouter()

func NewUrlRouter() *Service {
	return &Service{url2method: map[string]Router{}}
}

func (s *Service) WithUrl2Proc(url string) websocket.Proc {
	return s.url2method[url]
}

func (s *Service) Blueprint(path string, args Packet, codec gnetws.Serialize) Blueprint {
	refType := reflect.TypeOf(args)
	s.url2method[path] = &blueprint{
		codex: codec,
		args: sync.Pool{
			New: func() interface{} {
				return reflect.New(refType.Elem()).Interface()
			},
		},
		functionPool:             map[string]sync.Pool{},
		blueprintFunctionPlugins: new(Plugins),
	}

	return s.url2method[path].(*blueprint)
}

func (s *Service) Route(path string, codec gnetws.Serialize, proc websocket.Handler) (route Router, err error) {
	if _, ok := s.url2method[path]; ok {
		err = fmt.Errorf("please note that, path [%s] has been overwritten", path)
	}

	refType := reflect.TypeOf(proc)
	s.url2method[path] = &router{
		codex: codec,
		args: sync.Pool{
			New: func() interface{} {
				return reflect.New(refType.Elem()).Interface()
			},
		},
		routerMethodPlugins: new(Plugins),
	}

	return s.url2method[path], err
}
