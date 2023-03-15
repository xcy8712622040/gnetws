package serverhandler

import (
	"context"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/xcy8712622040/gnetws/cron-3"
	"time"
)

const (
	Conn          = "__conn__"
	HandlerServer = "__Handler_server__"
)

type Option func(s *ServerHandler)

func WithLogger(log logging.Logger) Option {
	return func(s *ServerHandler) {
		s.Logger = log
	}
}

func WithCronOnTicker(c *cron.Cron) Option {
	return func(s *ServerHandler) {
		s.Cron = c
	}
}

type ServerHandler struct {
	cron.EventCron
	logging.Logger

	addr    string
	Plugins plugins

	ctx context.Context
}

func NewHandler(opts ...Option) *ServerHandler {
	handler := new(ServerHandler)

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

func (s *ServerHandler) OnShutdown(eng gnet.Engine) {
	s.Plugins.EventServerOnShutdown(eng)
}

func (s *ServerHandler) OnBoot(eng gnet.Engine) (action gnet.Action) {
	go func() {
		<-s.ctx.Done()
		s.Warnf("server handler exit. cause: ", s.ctx.Err())

		func(err error) {
			if err != nil {
				s.Errorf("server handler stop error: %s", err.Error())
			}
		}(eng.Stop(context.Background()))
	}()

	if s.EventCron.Cron != nil {
		s.EventCron.Init()
	}
	return s.Plugins.EventServerOnBoot(eng)
}

func (s *ServerHandler) OnTick() (delay time.Duration, action gnet.Action) {
	return s.EventCron.Ticker()
}

func (s *ServerHandler) OnOpen(conn gnet.Conn) (out []byte, action gnet.Action) {
	ctx := WithContext(
		s.ctx,
		s.Logger,
		[2]interface{}{Conn, conn},
		[2]interface{}{HandlerServer, s},
	)

	conn.SetContext(ctx)

	s.Debugf(
		"OnOpen: client [%s] -> server handler [%s]",
		conn.RemoteAddr().String(), conn.LocalAddr().String(),
	)

	return s.Plugins.EventServerOnOpen(ctx)
}

func (s *ServerHandler) OnClose(conn gnet.Conn, err error) (action gnet.Action) {
	var ctx *Context

	if v := conn.Context(); ctx != nil {
		defer v.(*Context).Close()
	}

	s.Debugf(
		"OnClose: client [%s] -> server handler [%s] ERROR:[%s]",
		conn.RemoteAddr().String(), conn.LocalAddr().String(), err,
	)

	return s.Plugins.EventServerOnClose(ctx, err)
}

func (s *ServerHandler) OnTraffic(conn gnet.Conn) (action gnet.Action) {
	ctx := conn.Context().(*Context)

	if action = s.Plugins.EventServerOnTrafficPre(ctx, conn.InboundBuffered()); action != gnet.None {
		return action
	}

	if err := ctx.WithConn(conn); err != nil {
		action = gnet.Close
		s.Errorf("OnTraffic: client [%s]: %s", conn.RemoteAddr().String(), err)
	}

	return action
}

func (s *ServerHandler) Start(ctx context.Context, addr string, opts ...gnet.Option) error {
	s.ctx, s.addr = ctx, addr

	if s.Logger == nil {
		s.Logger = logrus.New()
	}

	return gnet.Run(s, addr, append([]gnet.Option{gnet.WithLogger(s.Logger)}, opts...)...)
}
