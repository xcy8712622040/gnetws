package main

import (
	"context"
	"flag"
	"github.com/panjf2000/gnet/v2"
	"github.com/sirupsen/logrus"
	"github.com/xcy8712622040/gnetws/cron-3"
	_ "github.com/xcy8712622040/gnetws/example-app/application"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket/dstservice"
	"github.com/xcy8712622040/gnetws/serverhandler"
	"log"
	"os/signal"
	"syscall"
)

var ProtoAddr = flag.String("addr", "tcp://:8080", "listener addr")

type WithOnUpgradePlugin struct{}

// OnUpgrade ws升级成功之后注册ws帧处理
func (w *WithOnUpgradePlugin) OnUpgrade(ctx *serverhandler.Context, conn *websocket.Conn, path string) error {
	if _, err := conn.WebSocketTextWriter().Write([]byte(`Success`)); err != nil {
		return err
	}
	WithFrame := &websocket.WithWebSocketFrameHandler{
		FrameHandler: dstservice.GlobalService.WithUrl2Proc(path),
	}
	ctx.WithHandler(WithFrame)
	return nil
}

type WithDefaultService struct{}

// OnOpen tcp链接建立成功之后 注册 http升websocket
func (w *WithDefaultService) OnOpen(ctx *serverhandler.Context) (out []byte, action gnet.Action) {
	WithUpHandler := new(websocket.WithWebSocketUpgradeHandle)
	WithUpHandler.Plugins().Add(new(WithOnUpgradePlugin))
	ctx.WithHandler(WithUpHandler)
	return out, action
}

func main() {
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(new(logrus.JSONFormatter))

	Serve := serverhandler.NewHandler(
		serverhandler.WithCron(cron.New(cron.WithSeconds())),
		serverhandler.WithLogger(logrus.WithField("kind", "test")),
	)

	Serve.Plugins().Add(new(WithDefaultService)) // 注册逻辑处理器

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, // Ctrl+C
	)

	//_, _ = Serve.Cron().AddFunc("*/1 * * * * *", func() {
	//	ms := runtime.MemStats{}
	//	runtime.ReadMemStats(&ms)
	//	logrus.Infof(
	//		"NumGoroutine:%d  MemAlloc:%dMB",
	//		runtime.NumGoroutine(),
	//		ms.Sys/1024/1024,
	//	)
	//})

	defer cancel()
	func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}(Serve.Start(ctx, *ProtoAddr, gnet.WithMulticore(true), gnet.WithTicker(true)))
}
