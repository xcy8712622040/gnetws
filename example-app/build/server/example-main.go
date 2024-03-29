package main

import (
	"context"
	"flag"
	"github.com/panjf2000/gnet/v2"
	"github.com/sirupsen/logrus"
	"github.com/xcy8712622040/gnetws/cron-3"
	_ "github.com/xcy8712622040/gnetws/example-app/application"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket/dstservice"
	"github.com/xcy8712622040/gnetws/serverhandler"
	"log"
	"os/signal"
	"runtime"
	"syscall"
)

var ProtoAddr = flag.String("addr", "tcp://:8080", "listener addr")

type WithDefaultService struct{}

func (w *WithDefaultService) OnOpen(ctx *serverhandler.Context) (out []byte, action gnet.Action) {
	ctx.WithHandler(dstservice.GlobalService)
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

	_, _ = Serve.Cron().AddFunc("*/1 * * * * *", func() {
		ms := runtime.MemStats{}
		runtime.ReadMemStats(&ms)
		logrus.Infof(
			"NumGoroutine:%d  MemAlloc:%dMB",
			runtime.NumGoroutine(),
			ms.Sys/1024/1024,
		)
	})

	defer cancel()
	func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}(Serve.Start(ctx, *ProtoAddr, gnet.WithMulticore(true), gnet.WithTicker(true)))
}
