/******************************
 * @Developer: many
 * @File: conn.go
 * @Time: 2022/6/10 11:30
******************************/

package eventserve

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

func (self *TextWriter) Flush() error {
	defer self.mutex.Unlock()
	return self.Writer.Flush()
}

func (self *TextWriter) Write(p []byte) (int, error) {
	self.mutex.Lock()
	return self.Writer.Write(p)
}

type Conn struct {
	gnet.Conn

	wait     int32
	wswriter *TextWriter
	wsreader *wsutil.Reader
	limitRdr *io.LimitedReader
}

func (self *Conn) Free() {
	if atomic.AddInt32(&self.wait, -1) <= 0 {
		defer pool.Put(self)
		if self.Conn != nil {
			self.Conn = nil
		}
	}
}

func (self *Conn) AwaitAdd() {
	atomic.AddInt32(&self.wait, 1)
}

func (self *Conn) State() ws.State {
	return self.wsreader.State
}

func (self *Conn) FrameReader() io.Reader {
	return self.wsreader
}

func (self *Conn) WebSocketTextWriter() *TextWriter {
	return self.wswriter
}

func (self *Conn) NextFrame() (err error) {
	var fn int64
	var head ws.Header
	for {
		fn, err = DeFrameLength(self.Conn)
		switch err {
		case nil, io.ErrShortBuffer:
		default:
			return err
		}

		if fn <= 0 {
			return io.EOF
		}

		if int64(self.Conn.InboundBuffered()) < fn {
			return io.EOF
		}

		self.limitRdr.N += fn

		head, err = self.wsreader.NextFrame()
		if err != nil {
			return err
		}

		if !head.OpCode.IsControl() {
			return err
		}

		if err = wsutil.ControlFrameHandler(self.Conn, ws.StateServerSide)(head, self.FrameReader()); err != nil {
			return err
		}
	}
}

func FrameConvert(conn gnet.Conn) *Conn {
	var frameconn *Conn

	if c := pool.Get(); c != nil {
		frameconn = c.(*Conn)

		frameconn.wait = 1
		frameconn.Conn = conn
		frameconn.limitRdr.N = 0
	} else {
		frameconn = &Conn{
			wait: 1,
			Conn: conn,
		}
		frameconn.limitRdr = &io.LimitedReader{R: frameconn}
		frameconn.wsreader = wsutil.NewReader(frameconn.limitRdr, ws.StateServerSide)
		frameconn.wswriter = &TextWriter{Writer: *wsutil.NewWriter(frameconn, ws.StateServerSide, ws.OpText)}
	}

	return frameconn
}
