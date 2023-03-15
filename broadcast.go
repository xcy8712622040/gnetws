package gnetws

import (
	"bytes"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/panjf2000/gnet/v2"
	"github.com/sirupsen/logrus"
	"io"
	"sync"
)

type Buffer struct {
	sync.Mutex
	bytes.Buffer
}

func (b *Buffer) Read(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()
	return b.Buffer.Read(p)
}

func (b *Buffer) Write(buf []byte) (int, error) {
	b.Lock()
	defer b.Unlock()
	return b.Buffer.Write(buf)
}

type Packing interface {
	io.Writer
	Flush() error
	SubPackage() (p []byte)
}

type Broadcast struct {
	pack  Packing
	codec Serialize

	storage sync.Map
	running chan struct{}
}

func NewBroadcast(pack Packing, codec Serialize) *Broadcast {
	return &Broadcast{
		pack:    pack,
		codec:   codec,
		storage: sync.Map{},
		running: make(chan struct{}, 1),
	}
}

func (b *Broadcast) WriteOffConn(c gnet.Conn) {
	b.storage.Delete(c)
}

func (b *Broadcast) RegisterConn(c gnet.Conn) {
	b.storage.Store(c, struct{}{})
}

func (b *Broadcast) SendMessage(x interface{}) (err error) {
	if err = b.codec.NewEnCodec(b.pack).Encode(x); err != nil {
		return err
	}
	if err = b.pack.Flush(); err != nil {
		return err
	}
	select {
	case b.running <- struct{}{}:
		go b.emit()
	default:
		logrus.Debugf("broadcast is running")
	}
	return
}

func (b *Broadcast) emit() {
	defer func() { <-b.running }()
	for pk := b.pack.SubPackage(); pk != nil; pk = b.pack.SubPackage() {
		b.storage.Range(func(key, value interface{}) bool {
			_ = key.(gnet.Conn).AsyncWrite(pk, nil)
			return true
		})
	}
}

type WebSocketWrapper struct {
	packet []byte
	buffer *Buffer
	writer *wsutil.Writer
}

func NewWebSocketWrapper() *WebSocketWrapper {
	buf := new(Buffer)
	return &WebSocketWrapper{
		buffer: buf,
		packet: make([]byte, 0, 4096),
		writer: wsutil.NewWriter(buf, ws.StateServerSide, ws.OpText),
	}
}

func (w *WebSocketWrapper) Flush() error {
	return w.writer.Flush()
}

func (w *WebSocketWrapper) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

func (w *WebSocketWrapper) SubPackage() []byte {
	n := bytes.IndexByte(w.buffer.Bytes(), '\n')
	if n <= 0 {
		return nil
	}

	if n <= cap(w.packet) {
		w.packet = w.packet[:n]
	} else {
		w.packet = make([]byte, n, n)
	}

	if _, err := w.buffer.Read(w.packet); err != nil {
		return nil
	}

	if len(w.packet) > 0 && w.packet[len(w.packet)-1] == '\r' {
		w.packet = w.packet[:len(w.packet)-1]
	}

	return w.packet
}
