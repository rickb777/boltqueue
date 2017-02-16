// Copyright (c) 2015 Andy Walker & Rick Beton
// Use of this source code is governed by the MIT License that can be
// found in the LICENSE file.

/*
Package boltqueue provides a persistent queue or priority queue based on
boltdb (https://github.com/boltdb/bolt)

Priority Queue

The PQueue type represents a priority queue. Messages may be
inserted into the queue at a numeric priority. Higher numbered priorities
take precedence over lower numbered ones.
Messages are dequeued following priority order, then time
ordering, with the oldest messages of the highest priority emerging
first.

There is no practical limit on the number of priorities, but a smaller number
will typically give better performance than a larger number.

The IChan type represents an unbounded channel with one priority, backed
by a PQueue. As with ordinary channels, messages are inserted into one
end and received from the other end in the same order. The size of the
channel's buffer is limited only by space on the fiing system.

The sending end
*/
package boltqueue
