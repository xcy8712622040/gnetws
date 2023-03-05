package websocketcli

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"io"
	"net"
)

type WebsocketClient struct {
	conn   net.Conn
	reader *wsutil.Reader
	writer *wsutil.Writer
}

func (w *WebsocketClient) Close() error {
	return w.conn.Close()
}

func (w *WebsocketClient) Dial(ctx context.Context, url string) (err error) {
	w.conn, _, _, err = ws.Dial(ctx, url)
	if err != nil {
		return err
	}

	w.reader = wsutil.NewReader(w.conn, ws.StateClientSide)
	w.writer = wsutil.NewWriter(w.conn, ws.StateClientSide, ws.OpText)
	return
}

func (w *WebsocketClient) Send(p []byte) (n int, err error) {
	defer func() {
		if err == nil {
			err = w.writer.Flush()
		}
	}()

	n, err = w.writer.Write(p)

	return n, err
}

func (w *WebsocketClient) Recv(writer io.Writer) (int64, error) {
	var (
		err  error
		head ws.Header
	)
	for head, err = w.reader.NextFrame(); err == nil && head.OpCode.IsControl(); head, err = w.reader.NextFrame() {
		if err = wsutil.ControlFrameHandler(w.conn, ws.StateClientSide)(head, w.reader); err != nil {
			return 0, err
		}
	}

	if err != nil {
		return 0, err
	}

	return io.Copy(writer, w.reader)
}
