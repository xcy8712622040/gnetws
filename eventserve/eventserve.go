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
	GnetConn          = "__gnet_conn__"
	GnetHandlerServer = "__gent_Handler_server__"
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

	Plugins plugins
}

func NewHeanler(opts ...Option) *ServerHandler {
	handler := new(ServerHandler)

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

func (self *ServerHandler) OnShutdown(eng gnet.Engine) {
	self.Plugins.EventServerOnShutdown(eng)
}
func (self *ServerHandler) OnBoot(eng gnet.Engine) (action gnet.Action) {
	if self.EventCron.Cron != nil {
		self.EventCron.Init()
	}
	return self.Plugins.EventServerOnBoot(eng)
}

func (self *ServerHandler) OnTick() (delay time.Duration, action gnet.Action) {
	return self.EventCron.Ticker()
}

func (self *ServerHandler) OnOpen(conn gnet.Conn) (out []byte, action gnet.Action) {
	ctx := WithWebSocketContext(
		self.todoctx, self.logger,
		[2]interface{}{GnetConn, conn},
		[2]interface{}{GnetHandlerServer, self},
	)

	if self.factory != nil {
		ctx.DoProc = self.factory()
	}

	conn.SetContext(ctx)
	self.logger.Debugf(
		"OnOpen: client [%s] -> server [%s]", conn.RemoteAddr().String(), conn.LocalAddr().String(),
	)

	return self.Plugins.EventServerOnOpen(ctx)
}

func (self *ServerHandler) OnClose(conn gnet.Conn, err error) (action gnet.Action) {
	var ctx *GnetContext

	if v := conn.Context(); ctx != nil {
		ctx = v.(*GnetContext)
		defer ctx.Cancel()
	}

	self.logger.Debugf(
		"OnClose: client [%s] -> server [%s] ERROR:[%s]", conn.RemoteAddr().String(), conn.LocalAddr().String(), err,
	)

	return self.Plugins.EventServerOnClose(ctx, err)
}

func (self *ServerHandler) OnTraffic(conn gnet.Conn) (action gnet.Action) {
	ctx := conn.Context().(*GnetContext)

	if action = self.Plugins.EventServerOnTrafficPre(ctx, conn.InboundBuffered()); action != gnet.None {
		return action
	}

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
