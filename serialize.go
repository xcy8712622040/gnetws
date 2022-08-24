/******************************
 * @Developer: many
 * @File: code.go
 * @Time: 2022/5/5 14:53
******************************/

package gnetws

import (
	"io"
)

// Serialize 序列化器
type Serialize interface {
	NewEnCodec(w io.Writer) EnCode
	NewDeCodec(r io.Reader) DeCode
}

type DeCode interface {
	Decode(x interface{}) (err error)
}

type EnCode interface {
	Encode(x interface{}) (err error)
}
