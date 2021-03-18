package boltqueue

import (
	"encoding/binary"
	"sync/atomic"
	"time"
)

type atomicKey uint64

func newAtomicKey() *atomicKey {
	key := atomicKey(time.Now().UnixNano())
	return &key
}

func (a *atomicKey) get() uint64 {
	return atomic.AddUint64((*uint64)(a), 1)
}

func (a *atomicKey) GetBytes() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, a.get())
	return b
}
