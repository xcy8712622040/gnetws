package websocket

import (
	"github.com/gobwas/ws"
	"github.com/panjf2000/gnet/v2"
	"github.com/xcy8712622040/gnetws"
	"github.com/xcy8712622040/gnetws/serverhandler"
	"io"
)

const (
	RequestHeader = "__request_header__"
)

var (
	globalPlugins = new(Plugins)
)

type Proc interface {
	Args() (x interface{})
	Codex() gnetws.Serialize
	DoProc(ctx *serverhandler.Context, conn *Conn, x interface{})
}

type WithWebSocketFrameHandler struct {
	FrameHandler Proc
}

func (w *WithWebSocketFrameHandler) Plugins() *Plugins {
	return globalPlugins
}

func (w *WithWebSocketFrameHandler) WithConn(ctx *serverhandler.Context, conn gnet.Conn) error {
	frame := FrameConvert(conn)

	defer frame.Free()
	for err := frame.NextFrame(); err != io.EOF; err = frame.NextFrame() {
		if err != nil {
			w.Plugins().WsOnClosedPlugin(ctx, err)
			return err
		}

		x := w.FrameHandler.Args()
		if err = w.FrameHandler.Codex().NewDeCodec(frame.FrameReader()).Decode(x); err == nil {
			w.FrameHandler.DoProc(ctx, frame, x)
		} else {
			ctx.Logger().Errorf("DoProc Error: %s", err)
		}
	}

	return nil
}

type WithWebSocketUpgradeHandle struct{}

func (w *WithWebSocketUpgradeHandle) Plugins() *Plugins {
	return globalPlugins
}

func (w *WithWebSocketUpgradeHandle) WithConn(ctx *serverhandler.Context, conn gnet.Conn) error {
	up := ws.Upgrader{}

	url := ""
	header := map[string]string{}

	up.OnRequest = func(uri []byte) (err error) {
		url = string(uri)
		return
	}

	up.OnHeader = func(key, value []byte) error {
		header[string(key)] = string(value)
		return nil
	}

	if _, err := up.Upgrade(conn); err != nil {
		return err
	}
	ctx.MetaData().SetInterface(RequestHeader, header)

	return w.Plugins().WsOnUpgrade(ctx, conn, url)
}
