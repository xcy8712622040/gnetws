package websocket

import "github.com/xcy8712622040/gnetws/serverhandler"

type WsPlugin interface{}

type (
	// OnUpgradePlugin ws链接建立成功之后
	OnUpgradePlugin interface {
		OnUpgrade(ctx *serverhandler.Context) error
	}

	// OnCallPrePlugin 逻辑处理之前
	OnCallPrePlugin interface {
		OnCallPre(ctx *serverhandler.Context, x interface{}) error
	}

	// OnCallPostPlugin 逻辑处理之后
	OnCallPostPlugin interface {
		OnCallPost(ctx *serverhandler.Context, x interface{}, reply interface{}) error
	}
)
