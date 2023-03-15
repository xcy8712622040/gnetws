package websocket

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/panjf2000/gnet/v2"
	"io"
	"sync"
	"sync/atomic"
)

var pool = sync.Pool{}

type TextWriter struct {
	wsutil.Writer

	mutex sync.Mutex
}

func (t *TextWriter) Flush() error {
	defer t.mutex.Unlock()
	return t.Writer.Flush()
}

func (t *TextWriter) Write(p []byte) (int, error) {
	t.mutex.Lock()
	return t.Writer.Write(p)
}

type Conn struct {
	gnet.Conn

	cite     int32
	writer   *TextWriter
	reader   *wsutil.Reader
	limitRdr *io.LimitedReader
}

func (c *Conn) Free() {
	if atomic.AddInt32(&c.cite, -1) <= 0 {
		defer pool.Put(c)
		if c.Conn != nil {
			c.Conn = nil
		}
	}
}

func (c *Conn) AddCite() {
	atomic.AddInt32(&c.cite, 1)
}

func (c *Conn) State() ws.State {
	return c.reader.State
}

func (c *Conn) FrameReader() io.Reader {
	return c.reader
}

func (c *Conn) WebSocketTextWriter() *TextWriter {
	return c.writer
}

func (c *Conn) NextFrame() (err error) {
	var fn int64
	var head ws.Header
	for {
		fn, err = DeFrameLength(c.Conn)
		switch err {
		case nil, io.ErrShortBuffer:
		default:
			return err
		}

		if fn <= 0 {
			return io.EOF
		}

		if int64(c.Conn.InboundBuffered()) < fn {
			return io.EOF
		}

		c.limitRdr.N += fn

		head, err = c.reader.NextFrame()
		if err != nil {
			return err
		}

		if !head.OpCode.IsControl() {
			return err
		}

		if err = wsutil.ControlFrameHandler(c.Conn, ws.StateServerSide)(head, c.FrameReader()); err != nil {
			return err
		}
	}
}

func FrameConvert(conn gnet.Conn) *Conn {
	var frame *Conn

	if c := pool.Get(); c != nil {
		frame = c.(*Conn)

		frame.cite = 1
		frame.Conn = conn
		frame.limitRdr.N = 0
	} else {
		frame = &Conn{
			cite: 1,
			Conn: conn,
		}
		frame.limitRdr = &io.LimitedReader{R: frame}
		frame.reader = wsutil.NewReader(frame.limitRdr, ws.StateServerSide)
		frame.writer = &TextWriter{Writer: *wsutil.NewWriter(frame, ws.StateServerSide, ws.OpText)}
	}

	return frame
}
