/******************************
 * @Developer: many
 * @File: api.go
 * @Time: 2022/6/7 15:53
******************************/

package echo

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xcy8712622040/gnetws"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket/dstservice"
	"github.com/xcy8712622040/gnetws/serverhandler"
	"io"
)

type JsonCodec struct {
	r io.Reader
	w io.Writer
}

func (j *JsonCodec) NewEnCodec(w io.Writer) gnetws.Encode {
	j.w = w
	return j
}

func (j *JsonCodec) Encode(x interface{}) (err error) {
	_, err = j.w.Write([]byte(*x.(*Data)))
	return
}

func (j *JsonCodec) NewDeCodec(r io.Reader) gnetws.Decode {
	j.r = r
	return j
}

func (j *JsonCodec) Decode(x interface{}) (err error) {
	if buf, err := io.ReadAll(j.r); err != nil {
		return err
	} else {
		if d, ok := x.(*Data); ok {
			*d = Data(buf)
			fmt.Println("=============>", *d)
		} else {
			fmt.Println("转换错误！！！！！！")
		}
	}
	return err

}

type Data string

func (d *Data) Proc(ctx *serverhandler.Context) interface{} {
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
