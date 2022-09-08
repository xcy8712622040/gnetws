package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket/websocketcli"
	"io"
	"time"
)

func main() {
	cli := websocketcli.WebscoketClient{}
	_ = cli.Dial(context.Background(), "ws://172.24.168.155:8080/")

	var sendt int64
	read := bytes.NewBuffer(make([]byte, 0))
	ppt := time.NewTicker(time.Second)
	go func() {
		for {
			<-ppt.C
			fmt.Println(time.Now(), sendt)
		}
	}()

	go func() {
		var err error
		defer func() { fmt.Println("recv exit:", err) }()
		for _, err = cli.Recv(read); err == nil; _, err = cli.Recv(read) {
			if _, err := io.ReadAll(read); err != nil {
				break
			} else {
				sendt++
			}
		}
	}()

	var err error
	//ticker := time.NewTicker(10 * time.Microsecond)
	defer func() { fmt.Println("send exit:", err) }()
	for {
		//<-ticker.C
		if _, err = cli.Send([]byte(`{"head":"test", "data":{"a":"111111"}}`)); err != nil {
			break
		}
	}
}
