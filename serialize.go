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
	NewEnCodec(w io.Writer) Encode
	NewDeCodec(r io.Reader) Decode
}

type Decode interface {
	Decode(x interface{}) (err error)
}

type Encode interface {
	Encode(x interface{}) (err error)
}
