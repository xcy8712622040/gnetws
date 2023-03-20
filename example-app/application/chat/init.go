/******************************
 * @Developer: many
 * @File: init.go
 * @Time: 2022/6/14 9:02
******************************/

package chat

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/xcy8712622040/gnetws"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket/dstservice"
	"github.com/xcy8712622040/gnetws/serverhandler"
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

type PreDoProc struct{}

// OnCallPre 逻辑处理之前
func (p *PreDoProc) OnCallPre(ctx *serverhandler.Context, x interface{}) (err error) {
	d := x.(*Packet)
	ctx.Logger().Debugf("chat handler proc with pre:", d.Head(), d.Payload)
	return
}

// OnCallPost 逻辑处理之后
func (p *PreDoProc) OnCallPost(ctx *serverhandler.Context, x interface{}, reply interface{}) (err error) {
	d := x.(*Packet)
	switch d.Head() {
	case "to":
		r := reply.(*Data)
		ctx.Logger().Debugf("chat handler proc with post:", r.Head, r.Data)
	}
	return
}

var blueprint dstservice.Blueprint

func init() {
	blueprint = dstservice.GlobalService.Blueprint("/chat", new(Packet), new(JsonCodec))
	blueprint.Plugins().Add(new(PreDoProc))
	logrus.Info("chat blueprint[ /chat ] router [ to ]:", blueprint.Route("to", new(Data)))
}
