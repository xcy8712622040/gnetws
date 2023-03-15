/******************************
 * @Developer: many
 * @File: api.go
 * @Time: 2022/6/14 9:06
******************************/

package chat

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xcy8712622040/gnetws/serverhandler"
)

type Data struct {
	Head string            `json:"head"`
	Data map[string]string `json:"data"`
}

func (d *Data) Proc(ctx *serverhandler.Context) interface{} {
	fmt.Println(d)
	d.Data["result"] = "hello world"
	return d
}

func init() {
	logrus.Info("chat blueprint[ /chat ] router [ to ]:", router.Route("to", new(Data)))
}
