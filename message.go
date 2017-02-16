package boltqueue

import "fmt"

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
