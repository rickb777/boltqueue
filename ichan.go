package boltqueue

type ErrorHandler func(error)

type puller struct {
	pqueue    *PQueue
	eh        func(error)
	poke      <-chan struct{}
	queueSize int64
}

type IChan struct {
	pqueue *PQueue
	eh     func(error)
	poke   chan struct{}
	input  chan []byte
	output chan []byte
	puller *puller
}

// NewIChan creates a new file-backed infinite channel. It uses the specified
// filename to create a BoltDB database that implements the channel persistence.
// If the filename is a directory name ending with '/', a unique filename is generated and appended to it.
// The channel's buffer is limited only by space available on the filesystem.
func NewIChan(filename string) (*IChan, error) {
	q, err := NewPQueue(filename, 1)
	return NewIChanOf(q), err
}

// NewIChanOf creates a new file-backed infinite channel from a BoltDB database.
// The channel's buffer is limited only by space available on the filesystem.
func NewIChanOf(pq *PQueue) *IChan {
	poke := make(chan struct{})
	data := make(chan []byte)

	ichan := &IChan{pqueue: pq, poke: poke, output: data}

	puller := &puller{
		pqueue:    ichan.pqueue,
		poke:      ichan.poke,
		queueSize: ichan.pqueue.size,
	}
	ichan.puller = puller

	go puller.recv(data)

	ichan.poke <- struct{}{}
	return ichan
}

// SetErrorHandler registers a function to handle errors at the receiving end.
func (c *IChan) SetErrorHandler(eh func(error)) {
	c.eh = eh
	c.puller.eh = eh
}

// SendEnd gets the input end of the IChan. The first time this is called, a new goroutine
// is started that transfers messages into the IChan.
//
// It is safe to share this channel between several goroutines; when this is done, a pseudo-random
// selection is made between them and only one goroutine receives each message (this is normal
// Go behaviour).
//
// When you have finished, you muse close the channel (as is normal for Go channels), otherwise
// the resources will not be released cleanly.
//
// If you prefer for there not to be one extra goroutine and don't want the simple channel
// abstraction, don't use this method but instead use Send(), SendString() and then Close().
func (c *IChan) SendEnd() chan<- []byte {
	if c.input == nil {
		c.input = make(chan []byte)
		go func() {
			for v := range c.input {
				err := c.send(v)
				if err != nil && c.eh != nil {
					c.eh(err)
				}
			}
			c.doClose()
		}()
	}
	return c.input
}

// SendString sends a message via the channel.
// This is a direct function call unlike interacting with a channel end. Once SendEnd()
// has been used, you cannot then use this method too.
func (c *IChan) SendString(value string) error {
	return c.Send([]byte(value))
}

// Send sends a message via the channel.
// This is a direct function call unlike interacting with a channel end. Once SendEnd()
// has been used, you cannot then use this method too.
func (c *IChan) Send(value []byte) error {
	checkState(c.input)
	return c.send(value)
}

func (c *IChan) send(value []byte) error {
	err := c.pqueue.EnqueueValue(0, value)
	c.poke <- struct{}{}
	return err
}

// Close closes the channel and its underlying queue.
// This is a direct function call unlike interacting with a channel end. Once SendEnd()
// has been used, you cannot then use this method too. You need instead to close the channel
// returned by SendEnd().
func (c *IChan) Close() error {
	checkState(c.input)
	return c.doClose()
}

// Close closes the channel and its underlying queue.
func (c *IChan) doClose() error {
	close(c.poke)
	return nil
}

func checkState(c chan []byte) {
	if c != nil {
		panic("Invalid usage: the sending-end channel has been created. " +
			"Send and Close should not now be used directly.")
	}
}

//-------------------------------------------------------------------------------------------------

func (p *puller) sendOn(value []byte, ch chan<- []byte) bool {
	select {
	case _, ok := (<-p.poke):
		ch <- value
		if ok {
			p.queueSize++

		} else {
			return false // terminate
		}

	case ch <- value:
		// sent ok
	}
	return true
}

func (p *puller) deq(ch chan<- []byte) bool {
	//fmt.Printf("DequeueValue...\n")
	value, err := p.pqueue.DequeueValue()
	if err != nil {
		if p.eh != nil {
			p.eh(err)
		}

	} else if value != nil {
		if p.queueSize > 0 {
			p.queueSize--
		}
		//fmt.Printf("DequeueValue '%s' %d\n", string(value), p.queueSize)
		return p.sendOn(value, ch)
	}

	_, ok := <-p.poke
	if ok {
		p.queueSize++
		return true // keep going

	} else {
		return false // terminate
	}
}

func (p *puller) recv(ch chan<- []byte) {
	<-p.poke
	running := true
	for running {
		running = p.deq(ch)
	}
	close(ch)
	p.pqueue.Close()
}

// ReceiveEnd gets the output end of the IChan. The result is the channel end, not the messages.
// This channel end should be used repeatedly until the channel is closed.
//
// It is safe to share this channel between several goroutines; when this is done, a pseudo-random
// selection is made between them and only one goroutine receives each message (this is normal
// Go behaviour).
func (c *IChan) ReceiveEnd() <-chan []byte {
	return c.output
}
