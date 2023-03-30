package websocket

import (
	"bytes"
	"github.com/gobwas/ws"
	"github.com/panjf2000/gnet/v2"
	"github.com/xcy8712622040/gnetws"
	"github.com/xcy8712622040/gnetws/serverhandler"
	"io"
	"net/url"
)

const (
	Protocol     = "__protocol__"
	RequestQuery = "__request_query__"
)

var (
	globalPlugins = new(Plugins)
)

func WsPlugins() *Plugins {
	return globalPlugins
}

type Proc interface {
	Args() (x interface{})
	Codex() gnetws.Serialize
	DoProc(ctx *serverhandler.Context, conn *Conn, x interface{})
}

type WithWebSocketFrameHandler struct {
	FrameHandler Proc
}

func (w *WithWebSocketFrameHandler) WithConn(ctx *serverhandler.Context, conn gnet.Conn) error {
	frame := FrameConvert(conn)

	defer frame.Free()
	for err := frame.NextFrame(); err != io.EOF; err = frame.NextFrame() {
		if err != nil {
			globalPlugins.WsOnClosedPlugin(ctx, err)
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

func (w *WithWebSocketUpgradeHandle) WithConn(ctx *serverhandler.Context, conn gnet.Conn) error {
	up := ws.Upgrader{}

	var (
		path  = ""
		query = map[string][]string{}
	)

	up.OnRequest = func(uri []byte) (err error) {
		URL := bytes.Split(uri, []byte("?"))

		path = string(URL[0])
		if query, err = url.ParseQuery(string(URL[1])); err != nil {
			return err
		} else {
			ctx.MetaData().SetInterface(RequestQuery, query)
		}
		return
	}

	if hs, err := up.Upgrade(conn); err != nil {
		return err
	} else {
		ctx.MetaData().SetInterface(Protocol, hs.Protocol)
	}

	return globalPlugins.WsOnUpgrade(ctx, FrameConvert(conn), path)
}
