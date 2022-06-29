/******************************
 * @Developer: many
 * @File: code.go
 * @Time: 2022/5/5 14:53
******************************/

package dstservice

import (
	"io"
)

type DoCodec interface {
	NewEnCodec(w io.Writer) EnCodec
	NewDeCodec(r io.Reader) DeCodec
}

type DeCodec interface {
	Decode(x interface{}) (err error)
}

type EnCodec interface {
	Encode(x interface{}) (err error)
}
