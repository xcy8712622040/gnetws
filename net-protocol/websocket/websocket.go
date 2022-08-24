/******************************
 * @Developer: many
 * @File: websocket.go
 * @Time: 2022/6/9 16:28
******************************/

package websocket

import (
	"encoding/binary"
	"github.com/gobwas/ws"
	"github.com/panjf2000/gnet/v2"
)

func DeFrameLength(conn gnet.Conn) (int64, error) {
	var IntactFrameLength int64
	for {
		bts, err := conn.Peek(int(IntactFrameLength) + ws.MinHeaderSize)
		if err != nil {
			return 0, err
		}

		bts = bts[IntactFrameLength : int(IntactFrameLength)+ws.MinHeaderSize]

		var extra int
		var PayLoadLength int64

		FIN := bts[0]&0x80 != 0

		// MASK 掩码
		if bts[1]&0x80 != 0 {
			extra += 4
		}

		// Payload Length
		length := bts[1] & 0x7f

		switch {
		case length < 126:
			PayLoadLength = int64(length)
		case length == 126:
			extra += 2
		case length == 127:
			extra += 8
		default:
			return 0, ws.ErrHeaderLengthUnexpected
		}

		if extra == 0 {
			return IntactFrameLength + ws.MinHeaderSize + PayLoadLength, nil
		}

		bts, err = conn.Peek(int(IntactFrameLength) + ws.MinHeaderSize + extra)
		if err != nil {
			return 0, err
		}

		bts = bts[int(IntactFrameLength)+ws.MinHeaderSize : int(IntactFrameLength)+ws.MinHeaderSize+extra]
		switch {
		case length == 126:
			PayLoadLength = int64(binary.BigEndian.Uint16(bts[:2]))

		case length == 127:
			if bts[0]&0x80 != 0 {
				return 0, ws.ErrHeaderLengthMSB
			}
			PayLoadLength = int64(binary.BigEndian.Uint64(bts[:8]))
		}

		if !FIN {
			IntactFrameLength += int64(extra) + ws.MinHeaderSize + PayLoadLength
		} else {
			return IntactFrameLength + int64(extra) + ws.MinHeaderSize + PayLoadLength, nil
		}
	}
}
