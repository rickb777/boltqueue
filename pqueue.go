package boltqueue

import (
	"encoding/binary"
	"fmt"

	"github.com/boltdb/bolt"
	"os"
	"strings"
	"time"
)

// aKey singleton for assigning keys to messages
var aKey = new(atomicKey)

// PQueue is a priority queue backed by a Bolt database on disk
type PQueue struct {
	// When RetainOnClose is true, the database file will be preserved after Close() is called.
	// Normally, the file is deleted on Close().
	RetainOnClose bool

	conn        *bolt.DB
	size        int64
	maxPriority int64
}

// NewPQueue loads or creates a new PQueue with the given filename.
// If the filename is a directory name ending with '/', a unique filename is generated and appended to it.
// Specify the required range of priorities; available priorities are from 0 (lowest) to
// the specified number minus one.
func NewPQueue(filename string, priorities uint) (*PQueue, error) {
	if strings.HasSuffix(filename, "/") {
		filename = fmt.Sprintf("%spq%d.db", filename, time.Now().UnixNano())
	}
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, err
	}
	return WrapDB(db, priorities)
}

// WrapDB wraps an existing BoltDB.
// Specify the required range of priorities; available priorities are from 0 (lowest) to
// the specified number minus one.
func WrapDB(db *bolt.DB, priorities uint) (*PQueue, error) {
	q := &PQueue{false, db, 0, int64(priorities) - 1}
	var err error
	q.size, err = q.TotalSize()
	return q, err
}

func (b *PQueue) enqueueMessage(priority uint, key []byte, message *Message) error {
	ipri := int64(priority)
	if ipri > b.maxPriority {
		return fmt.Errorf("Invalid priority %d on Enqueue", priority)
	}
	p := priBytes(ipri, b.maxPriority)

	err1 := b.conn.Update(func(tx *bolt.Tx) error {

		// Get bucket for this priority level
		pb, err2 := tx.CreateBucketIfNotExists(p)
		if err2 != nil {
			return err2
		}

		err2 = pb.Put(key, message.value)
		if err2 == nil {
			// note that Bolt Update provides a write lock already (no need for extra sync)
			b.size += 1
		}
		return err2
	})

	return err1
}

// Enqueue adds a message to the queue at a specified priority (0=lowest).
func (b *PQueue) Enqueue(priority uint, message *Message) error {
	return b.enqueueMessage(priority, aKey.GetBytes(), message)
}

// EnqueueValue adds a byte slice value to the queue at a specified priority (0=lowest).
func (b *PQueue) EnqueueValue(priority uint, value []byte) error {
	return b.enqueueMessage(priority, aKey.GetBytes(), WrapBytes(value))
}

// EnqueueString adds a string value to the queue at a specified priority (0=lowest).
func (b *PQueue) EnqueueString(priority uint, value string) error {
	return b.EnqueueValue(priority, []byte(value))
}

// Requeue adds a message back into the queue, keeping its precedence.
// If added at the same priority, it should be among the first to dequeue.
// If added at a different priority, it will dequeue before newer messages
// of that priority.
func (b *PQueue) Requeue(priority uint, message *Message) error {
	if message.key == nil {
		return fmt.Errorf("Cannot requeue a new message.")
	}
	return b.enqueueMessage(priority, message.key, message)
}

// Dequeue removes the oldest, highest priority message from the queue and returns it.
// If there are no messages available, nil, nil will be returned.
func (b *PQueue) Dequeue() (*Message, error) {
	var m *Message

	err1 := b.conn.Update(func(tx *bolt.Tx) error {

		for pri := b.maxPriority; pri >= 0; pri-- {
			bucket := tx.Bucket(priBytes(pri, b.maxPriority))

			if bucket != nil && bucket.Stats().KeyN > 0 {
				cur := bucket.Cursor()
				k, v := cur.First() //Should not be empty by definition
				m = &Message{priority: uint(pri), key: cloneBytes(k), value: cloneBytes(v)}

				// Remove message
				if err2 := cur.Delete(); err2 != nil {
					return err2
				}
				// note that Bolt Update provides a write lock already (no need for extra sync)
				b.size -= 1
				break
			}
		}

		return nil
	})

	return m, err1
}

// DequeueValue removes the oldest, highest priority message from the queue and returns its byte slice.
func (b *PQueue) DequeueValue() ([]byte, error) {
	m, err := b.Dequeue()
	if m == nil || err != nil {
		return nil, err
	}
	return m.value, nil
}

// DequeueString removes the oldest, highest priority message from the queue and returns its value as a string.
func (b *PQueue) DequeueString() (string, error) {
	m, err := b.Dequeue()
	if m == nil || err != nil {
		return "", err
	}
	return string(m.value), nil
}

// Size returns the number of entries of a given priority from 0 to 255 (0=highest).
func (b *PQueue) Size(priority uint) (int, error) {
	ipri := int64(priority)
	if ipri > b.maxPriority {
		return 0, fmt.Errorf("Invalid priority %d for Size()", priority)
	}

	tx, err := b.conn.Begin(false)
	if err != nil {
		return 0, err
	}

	bucket := tx.Bucket(priBytes(ipri, b.maxPriority))
	if bucket == nil {
		return 0, nil
	}

	count := bucket.Stats().KeyN
	tx.Rollback()

	return count, nil
}

// TotalSize sums the sizes of all the priority queues.
func (b *PQueue) TotalSize() (int64, error) {
	var size int64 = 0
	err := b.conn.View(func(tx *bolt.Tx) error {
		for pri := b.maxPriority; pri >= 0; pri-- {
			p := priBytes(pri, b.maxPriority)
			bucket := tx.Bucket(p)
			if bucket != nil {
				size += int64(bucket.Stats().KeyN)
			}
			//fmt.Printf("size after %d = %d\n", pri, size)
		}
		return nil
	})
	return size, err
}

// ApproxSize returns the sum of the sizes of all the priority queues, approximately. If the queue size is
// changing rapidly, this figure will be inaccurate. However, obtaining this value is very quick.
func (b *PQueue) ApproxSize() int64 {
	return b.size
}

// Close closes the queue database.
func (b *PQueue) Close() error {
	if !b.RetainOnClose {
		defer os.Remove(b.conn.Path())
	}
	return b.conn.Close()
}

func cloneBytes(v []byte) []byte {
	clone := make([]byte, len(v))
	copy(clone, v)
	return clone
}

func priBytes(priority, max int64) (b []byte) {
	if max <= 0x100 {
		b = make([]byte, 1)
		b[0] = byte(priority)
	} else if max <= 0x10000 {
		b = make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(priority))
	} else if max <= 0x100000000 {
		b = make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(priority))
	} else {
		b = make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(priority))
	}
	return
}
