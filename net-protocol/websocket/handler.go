package websocket

import "github.com/xcy8712622040/gnetws/serverhandler"

type Handler interface {
	Proc(ctx *serverhandler.Context) interface{}
}
