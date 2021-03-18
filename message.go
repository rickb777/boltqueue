package boltqueue

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// Message represents a message in the priority queue
type Message struct {
	key      []byte
	value    []byte
	priority uint
}

// NewMessagef generates a new priority queue message from a formatted string.
// Formatting is as per fmt.Sprintf.
func NewMessagef(format string, arg ...interface{}) *Message {
	return NewMessage(fmt.Sprintf(format, arg...))
}

// NewMessage generates a new priority queue message from a string.
func NewMessage(value string) *Message {
	return WrapBytes([]byte(value))
}

// NewGobMessage generates a new priority queue message from a value via gob
// encoding. Any error results in a panic (usually arising due to missing gob
// type registration).
func NewGobMessage(value interface{}) *Message {
	b := &bytes.Buffer{}
	err := gob.NewEncoder(b).Encode(value)
	if err != nil {
		panic(err)
	}
	return WrapBytes(b.Bytes())
}

// WrapBytes generates a new priority queue message.
// Do not modify the source value after submitting the message.
func WrapBytes(value []byte) *Message {
	return &Message{nil, value, 0}
}

// Priority returns the priority the message had in the queue.
func (m *Message) Priority() uint {
	return m.priority
}

// String outputs the string representation of the message's value.
func (m *Message) String() string {
	return string(m.value)
}

// Value returns the message's value./
// This is a mutable slice and you should not normally modify it.
func (m *Message) Value() []byte {
	return m.value
}

// GobValue returns the message's value using gob decoding.
// This unpacks the data from NewGobMessage.
func (m *Message) GobValue(v interface{}) error {
	return gob.NewDecoder(bytes.NewBuffer(m.value)).Decode(v)
}
