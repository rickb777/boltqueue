package boltqueue

import (
	"encoding/binary"
	"sync"
	"time"
)

type atomicKey struct {
	sync.Mutex
	key int64
}

func (a *atomicKey) get() uint64 {
	t := time.Now().UnixNano()

	a.Lock()
	defer a.Unlock()

	if t <= a.key {
		t = a.key + 1
	}
	a.key = t

	return uint64(t)
}

func (a *atomicKey) GetBytes() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, a.get())
	return b
}
