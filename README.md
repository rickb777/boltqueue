# boltqueue[![GoDoc](https://godoc.org/github.com/rickb777/boltqueue?status.svg)](https://godoc.org/github.com/rickb777/boltqueue)
--
    import "github.com/rickb777/boltqueue"

Package boltqueue provides a persistent priority queue (`PQueue`) based on BoltDB
(https://github.com/boltdb/bolt).

Built on this, a channel implementation is provided that uses the BoltDB backing store
to provide a channel with persistent buffering on a large scale. This is intended to
complement (not replace) standard message queuing technologies.


## Priority Queue

The PQueue type represents a priority queue. Messages may be
inserted into the queue at a numeric priority. Higher numbered priorities
take precedence over lower numbered ones.
Messages are dequeued following priority order, then time
ordering, with the oldest messages of the highest priority emerging
first.


## IChan

The IChan channel provides a communication pipe with the same semantics as normal
buffered Go channels. The message type is always `[]byte`, however. Buffering uses
a BoltDB file store. This allows the channel to be persistent and outside-of-memory,
but will impair performance compared to an equivalent in-memory channel.
