package serverhandler

import (
	"context"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	Conn          = "__conn__"
	HandlerServer = "__Handler_server__"
)

type Option func(s *Handler)

func WithCron(c *cron.Cron) Option {
	return func(s *Handler) {
		s.Cron = c
	}
}

func WithLogger(log logging.Logger) Option {
	return func(s *Handler) {
		s.logger = log
	}
}

type Handler struct {
	*cron.Cron

	plugins Plugins

	engine gnet.Engine

	logger logging.Logger

	context    context.Context
	cancelFunc context.CancelFunc
}

func NewHandler(opts ...Option) *Handler {
	handler := new(Handler)

	for _, opt := range opts {
		opt(handler)
	}

	handler.context, handler.cancelFunc = context.WithCancel(context.TODO())

	return handler
}

func (h *Handler) Plugins() *Plugins { return &h.plugins }

func (h *Handler) OnShutdown(eng gnet.Engine) {
	defer h.cancelFunc()
	h.plugins.EventServerOnShutdown(eng)
}

func (h *Handler) OnBoot(eng gnet.Engine) (action gnet.Action) {
	h.engine = eng
	return h.plugins.EventServerOnBoot(eng)
}

func (h *Handler) OnTick() (delay time.Duration, action gnet.Action) {
	return h.plugins.EventServerOnTicker(h)
}

func (h *Handler) OnOpen(conn gnet.Conn) (out []byte, action gnet.Action) {
	ctx := WithContext(
		h.context, h.logger,
		[2]interface{}{Conn, conn},
		[2]interface{}{HandlerServer, h},
	)

	conn.SetContext(ctx)

	h.logger.Debugf(
		"OnOpen: client [%s] -> server handler [%s]",
		conn.RemoteAddr().String(), conn.LocalAddr().String(),
	)

	return h.plugins.EventServerOnOpen(ctx)
}

func (h *Handler) OnClose(conn gnet.Conn, err error) (action gnet.Action) {
	var ctx *Context

	if v := conn.Context(); ctx != nil {
		defer func() {
			ctx.Close()
		}()
		ctx = v.(*Context)
	}

	h.logger.Debugf(
		"OnClose: client [%s] -> server [%s] ERROR:[%s]",
		conn.RemoteAddr().String(), conn.LocalAddr().String(), err,
	)

	return h.plugins.EventServerOnClose(ctx, err)
}

func (h *Handler) OnTraffic(conn gnet.Conn) (action gnet.Action) {
	ctx := conn.Context().(*Context)

	if action = h.plugins.EventServerOnTrafficPre(ctx, conn.InboundBuffered()); action != gnet.None {
		return action
	}

	if err := ctx.DealWithConn(conn); err != nil {
		action = gnet.Close
		h.logger.Errorf("OnTraffic: client [%s]: %s", conn.RemoteAddr().String(), err)
	}

	return action
}

func (h *Handler) Start(ctx context.Context, addr string, opts ...gnet.Option) error {
	if h.logger == nil {
		h.logger = logrus.New()
	}

	if h.Cron != nil {
		h.Cron.Start()
	}

	go func() {
		<-ctx.Done()
		h.logger.Warnf("server handler exit. cause: ", ctx.Err())

		if h.Cron != nil {
			<-h.Cron.Stop().Done()
		}

		func(err error) {
			if err == nil {
				h.logger.Debugf("server handler normal shutdown")
			} else {
				h.logger.Errorf("server handler stop error: %s", err.Error())
			}
		}(h.engine.Stop(context.Background()))
	}()

	return gnet.Run(h, addr, append([]gnet.Option{gnet.WithLogger(h.logger)}, opts...)...)
}
