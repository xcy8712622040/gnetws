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

func (self *Buffer) Read(p []byte) (int, error) {
	self.Lock()
	defer self.Unlock()
	return self.Buffer.Read(p)
}

func (self *Buffer) Write(b []byte) (int, error) {
	self.Lock()
	defer self.Unlock()
	return self.Buffer.Write(b)
}

type Packing interface {
	io.Writer
	Flush() error
	Recovery(p []byte)
	SubPackage() (p []byte)
}

type Broadcast struct {
	pack  Packing
	codec DoCodec

	storage sync.Map
	running chan struct{}
}

func NewBroadcast(wrapper Packing, docodec DoCodec) *Broadcast {
	return &Broadcast{
		pack:    wrapper,
		codec:   docodec,
		storage: sync.Map{},
		running: make(chan struct{}, 1),
	}
}

func (self *Broadcast) WriteOffConn(c gnet.Conn) {
	self.storage.Delete(c)
}

func (self *Broadcast) RegisterConn(c gnet.Conn) {
	self.storage.Store(c, struct{}{})
}

func (self *Broadcast) SendMessage(x any) (err error) {
	if err = self.codec.NewEnCodec(self.pack).Encode(x); err != nil {
		return err
	}
	if err = self.pack.Flush(); err != nil {
		return err
	}
	select {
	case self.running <- struct{}{}:
		go self.emit()
	default:
		logrus.Debugf("broadcast is running")
	}
	return
}

func (self *Broadcast) emit() {
	defer func() { <-self.running }()
	for pk := self.pack.SubPackage(); pk != nil; pk = self.pack.SubPackage() {
		self.storage.Range(func(key, value any) bool {
			_ = key.(gnet.Conn).AsyncWrite(pk, nil)
			return true
		})
		self.pack.Recovery(pk) // 回收
	}
}

var ppool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024)
	},
}

func scale(b []byte, l int) []byte {
	if cap(b) >= l {
		return b[:l]
	}
	kb := func(i int) int {
		t := 1
		for t*1024 < i {
			t++
		}
		return t * 1024
	}

	return make([]byte, l, kb(l))
}

type WebSocketWrapper struct {
	buffer  *Buffer
	writetx *wsutil.Writer
}

func NewWebSocketWrapper() *WebSocketWrapper {
	buf := new(Buffer)
	return &WebSocketWrapper{
		buffer:  buf,
		writetx: wsutil.NewWriter(buf, ws.StateServerSide, ws.OpText),
	}
}

func (self *WebSocketWrapper) Recovery(p []byte) {
	ppool.Put(p)
}

func (self *WebSocketWrapper) Flush() error {
	return self.writetx.Flush()
}

func (self *WebSocketWrapper) Write(p []byte) (int, error) {
	return self.writetx.Write(p)
}

func (self *WebSocketWrapper) SubPackage() (p []byte) {
	n := bytes.IndexByte(self.buffer.Bytes(), '\n')
	if n <= 0 {
		return nil
	}
	p = ppool.Get().([]byte)

	scale(p, n) // 检测容量

	if _, err := self.buffer.Read(p); err != nil {
		return nil
	}

	if len(p) > 0 && p[len(p)-1] == '\r' {
		p = p[:len(p)-1]
	}

	return
}
