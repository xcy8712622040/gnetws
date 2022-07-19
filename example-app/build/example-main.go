package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/panjf2000/gnet/v2"
	"github.com/sirupsen/logrus"
	"github.com/xcy8712622040/gnetws/cron-3"
	"github.com/xcy8712622040/gnetws/dstservice"
	"github.com/xcy8712622040/gnetws/eventserve"
	_ "github.com/xcy8712622040/gnetws/example-app/application"
	"log"
	"os/signal"
	"syscall"
	"time"
)

var ProtoAddr = flag.String("addr", "tcp://:8080", "listener addr")

func main() {
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(new(logrus.JSONFormatter))

	Serve := eventserve.NewHeanler(
		eventserve.WithCronOnTicker(cron.New(cron.WithSeconds())),
		eventserve.WithLogger(logrus.WithField("kind", "test")),
		eventserve.WithHandlerFactory(func() eventserve.Handler {
			return dstservice.Handler
		}),
	)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, // kill
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGILL,  // 非法指令
		syscall.SIGABRT, // abort函数触发
	)

	_, _ = Serve.AddFunc("*/1 * * * * *", func() {
		fmt.Println("1", time.Now().Format("2006-01-02 15:04:05"))
	})

	_, _ = Serve.AddFunc("*/20 * * * * *", func() {
		fmt.Println("2", time.Now().Format("2006-01-02 15:04:05"))
	})

	_, _ = Serve.AddFunc("0 */1 * * * *", func() {
		fmt.Println("3", time.Now().Format("2006-01-02 15:04:05"))
	})

	_, _ = Serve.AddFunc("0 0 */1 * * *", func() {
		fmt.Println("4", time.Now().Format("2006-01-02 15:04:05"))
	})

	defer cancel()
	func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}(Serve.Start(ctx, *ProtoAddr, gnet.WithMulticore(true), gnet.WithTicker(true)))
}
