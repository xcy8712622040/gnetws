package main

import (
	"context"
	"flag"
	"fmt"
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
	"time"
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
		serverhandler.WithCronOnTicker(cron.New(cron.WithSeconds())),
		serverhandler.WithLogger(logrus.WithField("kind", "test")),
	)

	Serve.Plugins.Add(new(WithDefaultService)) // 注册逻辑处理器

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, // Ctrl+C
	)

	_, _ = Serve.AddFunc("*/1 * * * * *", func() {
		logrus.Info("NumGoroutine:", runtime.NumGoroutine())
		fmt.Println("1", time.Now().Format("2006-01-02 15:04:05"))
	})

	defer cancel()
	func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}(Serve.Start(ctx, *ProtoAddr, gnet.WithMulticore(true), gnet.WithTicker(true)))
}
