/******************************
 * @Developer: many
 * @File: context.go
 * @Time: 2022/5/24 15:12
******************************/

package serverhandler

import (
	"context"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"io"
)

const (
	METADATA = "__mate_data__"
)

type Proc interface {
	WithConn(ctx *Context, conn gnet.Conn) error
}

// Context 上下文
type Context struct {
	Proc
	context.Context

	cancel func()
	logger logging.Logger

	running chan struct{}
}

// Close 关闭
func (c *Context) Close() { c.cancel() }

// Logger 日志
func (c *Context) Logger() logging.Logger { return c.logger }

// MetaData 上下文元数据
func (c *Context) MetaData() *MateData { return c.Value(METADATA).(*MateData) }

// WithHandler 设置一个处理器
func (c *Context) WithHandler(proc Proc) { c.Proc = proc }

// WithConn 处理一个conn
func (c *Context) WithConn(conn gnet.Conn) error {
	if c.Proc == nil {
		if _, err := io.Copy(conn, conn); err != nil {
			return err
		}
	} else {
		return c.Proc.WithConn(c, conn)
	}

	return nil
}

// WithContext 创建一个上下文
func WithContext(ctx context.Context, logger logging.Logger, data ...[2]interface{}) *Context {
	meta := &MateData{
		data: map[interface{}]interface{}{},
	}

	for idx := range data {
		meta.SetInterface(data[idx][0], data[idx][1])
	}

	ctx, cancel := context.WithCancel(
		context.WithValue(ctx, METADATA, meta),
	)

	return &Context{logger: logger, Context: ctx, cancel: cancel, running: make(chan struct{}, 1)}
}
