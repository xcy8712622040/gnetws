/******************************
 * @Developer: many
 * @File: context.go
 * @Time: 2022/5/24 15:12
******************************/

package eventserve

import (
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"golang.org/x/net/context"
)

const (
	MATEDATA = "__mate_data__"
)

type Handler interface {
	Proc(ctx *GnetContext, conn gnet.Conn) error
}

type GnetContext struct {
	context.Context

	Cancel func()
	DoProc Handler
	Logger logging.Logger
}

func WithWebSocketContext(ctx context.Context, logger logging.Logger, mate ...[2]interface{}) *GnetContext {
	matedata := map[interface{}]interface{}{}
	for idx := range mate {
		matedata[mate[idx][0]] = mate[idx][1]
	}

	cancelctx, cancel := context.WithCancel(
		context.WithValue(
			ctx, MATEDATA, matedata,
		),
	)

	return &GnetContext{Context: cancelctx, Cancel: cancel, Logger: logger}
}
