/******************************
 * @Developer: many
 * @File: wsclient.go
 * @Time: 2022/6/22 9:33
******************************/

package websocketcli

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"io"
	"net"
)

type WebscoketClient struct {
	conn   net.Conn
	reader *wsutil.Reader
	writer *wsutil.Writer
}

func (self *WebscoketClient) Dial(ctx context.Context, url string) (err error) {
	self.conn, _, _, err = ws.Dial(ctx, url)
	if err != nil {
		return err
	}

	self.reader = wsutil.NewReader(self.conn, ws.StateClientSide)
	self.writer = wsutil.NewWriter(self.conn, ws.StateClientSide, ws.OpText)
	return
}

func (self *WebscoketClient) Send(p []byte) (n int, err error) {
	defer func() {
		if err == nil {
			err = self.writer.Flush()
		}
	}()

	n, err = self.writer.Write(p)

	return n, err
}

func (self *WebscoketClient) Recv(w io.Writer) (int64, error) {
	var (
		err  error
		head ws.Header
	)
	for head, err = self.reader.NextFrame(); err == nil && head.OpCode.IsControl(); head, err = self.reader.NextFrame() {
		if err = wsutil.ControlFrameHandler(self.conn, ws.StateClientSide)(head, self.reader); err != nil {
			return 0, err
		}
	}

	if err != nil {
		return 0, err
	}

	return io.Copy(w, self.reader)
}
