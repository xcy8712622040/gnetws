/******************************
 * @Developer: many
 * @File: init.go
 * @Time: 2022/6/14 9:02
******************************/

package chat

import (
	"encoding/json"
	"github.com/xcy8712622040/gnetws"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket/dstservice"
	"io"
)

type JsonCodec struct{}

func (self JsonCodec) NewEnCodec(w io.Writer) gnetws.EnCode {
	return json.NewEncoder(w)
}

func (self JsonCodec) NewDeCodec(r io.Reader) gnetws.DeCode {
	return json.NewDecoder(r)
}

type Packet struct {
	MsgHead string `json:"head"`
	Payload string `json:"payload"`
}

func (self *Packet) Head() string {
	return self.MsgHead
}

func (self *Packet) Read(p []byte) (int, error) {
	n := copy(p, self.Payload)
	return n, nil
}

var router = dstservice.Handler.Blueprint("/chat", new(Packet), new(JsonCodec))
