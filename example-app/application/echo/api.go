/******************************
 * @Developer: many
 * @File: api.go
 * @Time: 2022/6/7 15:53
******************************/

package echo

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/xcy8712622040/gnetws"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket/dstservice"
	"github.com/xcy8712622040/gnetws/serverhandler"
	"io"
	"strconv"
	"time"
)

type JsonCodec struct{}

func (j JsonCodec) NewEnCodec(w io.Writer) gnetws.Encode {
	return json.NewEncoder(w)
}

func (j JsonCodec) NewDeCodec(r io.Reader) gnetws.Decode {
	return json.NewDecoder(r)
}

type Data struct {
	Head string            `json:"head"`
	Data map[string]string `json:"data"`
}

func (d *Data) Proc(ctx *serverhandler.Context) interface{} {
	d.Data["result"] = strconv.Itoa(int(time.Now().UnixNano()))
	return d
}

type PreDoProc struct{}

// OnCallPre 逻辑处理之前
func (p *PreDoProc) OnCallPre(ctx *serverhandler.Context, x interface{}) (err error) {
	d := x.(*Data)
	ctx.Logger().Debugf("echo handler proc with pre:[%#v]", d)
	return
}

// OnCallPost 逻辑处理之后
func (p *PreDoProc) OnCallPost(ctx *serverhandler.Context, x interface{}, reply interface{}) (err error) {
	r := reply.(*Data)
	ctx.Logger().Debugf("echo handler proc with post:[%#v]", r)
	return
}

func init() {
	router, err := dstservice.GlobalService.Route(
		"/", new(JsonCodec), new(Data),
	)

	router.Plugins().Add(new(PreDoProc))
	logrus.Info("echo Handler [ / ] router:", err)
}
