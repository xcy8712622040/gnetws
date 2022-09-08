/******************************
 * @Developer: many
 * @File: context.go
 * @Time: 2022/5/24 15:12
******************************/

package eventserve

import (
	"context"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"sync"
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
	Locker sync.Locker
	Logger logging.Logger
}

// Value 获取上下文变量
func (c *GnetContext) Value(key interface{}) interface{} {
	if val := c.Context.Value(key); val != nil {
		return val
	}
	return c.Context.Value(MATEDATA).(map[interface{}]interface{})[key]
}

func WithWebSocketContext(ctx context.Context, logger logging.Logger, mate ...[2]interface{}) *GnetContext {
	matedata := &MateData{lock: sync.RWMutex{}, data: map[interface{}]interface{}{}}

	for idx := range mate {
		matedata.SetInterface(mate[idx][0], mate[idx][1])
	}

	cancelctx, cancel := context.WithCancel(
		context.WithValue(
			ctx, MATEDATA, matedata,
		),
	)

	return &GnetContext{Context: cancelctx, Cancel: cancel, Logger: logger, Locker: new(sync.Mutex)}
}
