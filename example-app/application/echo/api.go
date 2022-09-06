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
	"github.com/xcy8712622040/gnetws/eventserve"
	"github.com/xcy8712622040/gnetws/net-protocol/websocket/dstservice"
	"io"
	"strconv"
	"time"
)

type JsonCodec struct{}

func (self JsonCodec) NewEnCodec(w io.Writer) gnetws.Encode {
	return json.NewEncoder(w)
}

func (self JsonCodec) NewDeCodec(r io.Reader) gnetws.Decode {
	return json.NewDecoder(r)
}

type Data struct {
	Head string            `json:"head"`
	Data map[string]string `json:"data"`
}

func (self *Data) Proc(ctx *eventserve.GnetContext) interface{} {
	self.Data["resert"] = strconv.Itoa(int(time.Now().UnixNano()))
	return self
}

func init() {
	logrus.Info("echo Handler [ / ] router:", dstservice.Handler.Route("/", new(JsonCodec), new(Data)))
}
