package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket/websocketcli"
	"io"
	"log"
	"time"
)

func main() {
	cli := websocketcli.WebsocketClient{}
	if err := cli.Dial(context.Background(), "ws://127.0.0.1:8080/"); err != nil {
		log.Fatal(err)
	}

	var recv, send int64
	read := bytes.NewBuffer(make([]byte, 0))
	ppt := time.NewTicker(time.Second)
	go func() {
		for {
			<-ppt.C
			fmt.Println(fmt.Sprintf("%#v, send:%d, recv:%d", time.Now(), send, recv))
		}
	}()

	go func() {
		var err error
		defer func() { fmt.Println("receive exit:", err) }()
		for _, err = cli.Recv(read); err == nil; _, err = cli.Recv(read) {
			if _, err := io.ReadAll(read); err != nil {
				break
			} else {
				recv++
			}
		}
	}()

	var err error
	ticker := time.NewTicker(10 * time.Microsecond)
	defer func() { fmt.Println("send exit:", err) }()
	for {
		<-ticker.C
		if _, err = cli.Send([]byte(`{"head":"test", "data":{"a":"111111"}}`)); err != nil {
			break
		} else {
			send++
		}
	}
}
