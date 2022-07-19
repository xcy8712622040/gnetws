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
	"github.com/xcy8712622040/gnetws/dstservice"
	"github.com/xcy8712622040/gnetws/eventserve"
	"io"
	"strconv"
	"time"
)

type JsonCodec struct{}

func (self JsonCodec) NewEnCodec(w io.Writer) gnetws.EnCodec {
	return json.NewEncoder(w)
}

func (self JsonCodec) NewDeCodec(r io.Reader) gnetws.DeCodec {
	return json.NewDecoder(r)
}

type Data struct {
	Head string            `json:"head"`
	Data map[string]string `json:"data"`
}

func init() {
	logrus.Info("echo Handler [ / ] router:", dstservice.Handler.Route(
		"/", new(Data), new(JsonCodec),
		func(ctx *eventserve.WebSocketContext, args interface{}) interface{} {
			data := *args.(*Data)
			data.Data["resert"] = strconv.Itoa(int(time.Now().UnixNano()))
			return data
		},
	))
}
