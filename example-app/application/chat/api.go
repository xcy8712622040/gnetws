/******************************
 * @Developer: many
 * @File: api.go
 * @Time: 2022/6/14 9:06
******************************/

package chat

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xcy8712622040/gnetws/eventserve"
)

type Data struct {
	Head string            `json:"head"`
	Data map[string]string `json:"data"`
}

func init() {
	logrus.Info("chat blueprint[ /chat ] router [ to ]:", router.Route(
		"to", new(Data),
		func(ctx *eventserve.GnetContext, args interface{}) interface{} {
			data := *args.(*Data)
			fmt.Println(data)
			data.Data["resert"] = "hello world"
			return data
		},
	))
}
