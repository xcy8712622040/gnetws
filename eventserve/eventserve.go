/******************************
 * @Developer: many
 * @File: eventserve.go
 * @Time: 2022/5/24 15:13
******************************/

package eventserve

import (
	"context"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/xcy8712622040/gnetws/cron-3"
	"io"
	"time"
)

const (
	GnetConn        = "__gnet_conn__"
	WebSocketServer = "__ws_server__"
)

type Option func(s *ServerHandler)
type FactoryHandler func() Handler

func WithLogger(l logging.Logger) Option {
	return func(s *ServerHandler) {
		s.logger = l
	}
}

func WithCronOnTicker(c *cron.Cron) Option {
	return func(s *ServerHandler) {
		s.Cron = c
	}
}

func WithHandlerFactory(f FactoryHandler) Option {
	return func(s *ServerHandler) {
		s.factory = f
	}
}

type ServerHandler struct {
	cron.EventCron
	gnet.BuiltinEventEngine

	addr    string
	logger  logging.Logger
	factory FactoryHandler
	todoctx context.Context
}

func NewHeanler(opts ...Option) *ServerHandler {
	handler := new(ServerHandler)

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

func (self *ServerHandler) OnBoot(_ gnet.Engine) (action gnet.Action) {
	return self.EventCron.Init()
}

func (self *ServerHandler) OnTick() (delay time.Duration, action gnet.Action) {
	return self.EventCron.Ticker()
}
func (self *ServerHandler) OnOpen(conn gnet.Conn) (out []byte, action gnet.Action) {
	ctx := WithWebSocketContext(
		self.todoctx, self.logger,
		[2]interface{}{GnetConn, conn},
		[2]interface{}{WebSocketServer, self},
	)

	if self.factory != nil {
		ctx.DoProc = self.factory()
	}

	conn.SetContext(ctx)
	self.logger.Debugf(
		"OnOpen: client [%s] -> server [%s]", conn.RemoteAddr().String(), conn.LocalAddr().String(),
	)

	return
}

func (self *ServerHandler) OnClose(conn gnet.Conn, err error) (action gnet.Action) {
	if ctx := conn.Context(); ctx != nil {
		ctx.(*WebSocketContext).Cancel()
	}

	self.logger.Debugf(
		"OnClose: client [%s] -> server [%s] ERROR:[%s]", conn.RemoteAddr().String(), conn.LocalAddr().String(), err,
	)

	return
}

func (self *ServerHandler) OnTraffic(conn gnet.Conn) (action gnet.Action) {
	ctx := conn.Context().(*WebSocketContext)
	if ctx.DoProc != nil {
		if err := ctx.DoProc.Proc(ctx, conn); err != nil {
			action = gnet.Close
			self.logger.Errorf("OnTraffic: client [%s]: %s", conn.RemoteAddr().String(), err)
		}
	} else {
		if _, err := io.Copy(conn, conn); err != nil {
			action = gnet.Close
			self.logger.Errorf("OnTraffic: client [%s]: %s", conn.RemoteAddr().String(), err)
		}
	}
	return action
}

func (self *ServerHandler) Stop(err error) {
	self.logger.Warnf("server exit. cause: ", err)

	func(err error) {
		if err != nil {
			logrus.Println("gnet stop error: ", err.Error())
		}
	}(gnet.Stop(context.Background(), self.addr))
}

func (self *ServerHandler) Start(ctx context.Context, protoaddr string, opts ...gnet.Option) error {
	self.todoctx, self.addr = ctx, protoaddr

	go func() {
		<-ctx.Done()
		self.Stop(ctx.Err())
	}()

	if self.logger == nil {
		self.logger = logrus.New()
	}

	opts = append(opts, []gnet.Option{
		gnet.WithLogger(self.logger),
	}...)

	return gnet.Run(self, protoaddr, opts...)
}
