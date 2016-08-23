package utils

import (
	"sync"
)

const (
	tooBig = 1 << 30
)

type Buffer struct {
	Buff    []byte
	scratch [64]byte
}

func (b *Buffer) Reset() {
	if len(b.Buff) >= tooBig {
		b.Buff = b.scratch[0:0]
	} else {
		b.Buff = b.Buff[0:0]
	}
}

var BufferPool = sync.Pool{
	New: func() interface{} {
		b := new(Buffer)
		b.Buff = b.scratch[0:0]
		return b
	},
}

func (b *Buffer) Bytes() []byte {
	return b.Buff
}

// func GetBytes() []byte {
// 	buff := BufferPool.Get().(*Buffer)
// 	b := buff.Buff
// 	BufferPool.Put(buff)
// 	return b
// }

// func GetNByte(n int) []byte {
// 	b := GetBytes()
// 	if len(b) < n {
// 		t := make([]byte, n-len(b))
// 		b = append(b, t...)
// 	} else {
// 		b = b[0:n]
// 	}
// 	return b
// }

func GetBytes() []byte {
	return make([]byte, 10)
}

func GetNByte(n int) []byte {
	return make([]byte, n)
}
