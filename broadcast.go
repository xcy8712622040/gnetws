package gnetws

import (
	"bytes"
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
	Recycle(p []byte)
	SubPackage() (p []byte)
	SetRecipient(w io.ReadWriter)
}

type Broadcast struct {
	pack  Packing
	codec DoCodec

	storage sync.Map
	running chan struct{}
}

func NewBroadcast(wrapper Packing, docodec DoCodec) *Broadcast {
	buf := new(Buffer)
	wrapper.SetRecipient(buf)
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
		self.pack.Recycle(pk)
	}
}
