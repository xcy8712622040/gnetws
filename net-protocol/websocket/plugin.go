package websocket

import "github.com/xcy8712622040/gnetws/eventserve"

type WsPlugin interface{}

type (
	// ws链接建立成功之后
	OnUpgraderPlugin interface {
		OnUpgrader(ctx *eventserve.GnetContext) error
	}

	// 逻辑处理之前
	OnCallPrePlugin interface {
		OnCallPre(ctx *eventserve.GnetContext, x interface{}) error
	}

	// 逻辑处理之后
	OnCallPostPlugin interface {
		OnCallPost(ctx *eventserve.GnetContext, x interface{}, reply interface{}) error
	}
)
