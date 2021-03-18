# boltqueue

[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/rickb777/boltqueue)
[![Build Status](https://api.travis-ci.org/rickb777/boltqueue.svg?branch=master)](https://travis-ci.org/rickb777/boltqueue/builds)
[![Coverage Status](https://coveralls.io/repos/rickb777/boltqueue/badge.svg?branch=master&service=github)](https://coveralls.io/github/rickb777/boltqueue?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/rickb777/boltqueue)](https://goreportcard.com/report/github.com/rickb777/boltqueue)
[![Issues](https://img.shields.io/github/issues/rickb777/boltqueue.svg)](https://github.com/rickb777/boltqueue/issues)


    import "github.com/rickb777/boltqueue"

Package boltqueue provides a persistent priority queue (`PQueue`) using as its store 
[BBolt](https://pkg.go.dev/go.etcd.io/bbolt), which
is a derivative of [BoltDB](https://github.com/boltdb/bolt).

Built on this, a channel implementation is provided that uses the BBolt backing store
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


## Licence : MIT
