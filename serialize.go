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
