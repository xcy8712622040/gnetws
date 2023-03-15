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

func (j JsonCodec) NewEnCodec(w io.Writer) gnetws.Encode {
	return json.NewEncoder(w)
}

func (j JsonCodec) NewDeCodec(r io.Reader) gnetws.Decode {
	return json.NewDecoder(r)
}

type Packet struct {
	MsgHead string `json:"head"`
	Payload string `json:"payload"`
}

func (p *Packet) Head() string {
	return p.MsgHead
}

func (p *Packet) Read(pkg []byte) (int, error) {
	n := copy(pkg, p.Payload)
	return n, nil
}

var router = dstservice.GlobalService.Blueprint("/chat", new(Packet), new(JsonCodec))
