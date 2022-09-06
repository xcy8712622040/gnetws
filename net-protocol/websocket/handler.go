package websocket

import "github.com/xcy8712622040/gnetws/eventserve"

type Handler interface {
	Proc(ctx *eventserve.GnetContext) interface{}
}
