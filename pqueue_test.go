package boltqueue

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

var zero = uint(0)
var one = uint(1)
var five = uint(5)

func TestGobMessage(t *testing.T) {
	input := map[string]int{
		"a": 101,
		"b": 37,
	}
	m := NewGobMessage(input)

	output := make(map[string]int)
	err := m.GobValue(&output)
	if err != nil {
		t.Error(err)
	} else if len(output) != 2 {
		t.Errorf("Expected map size 2. Got: %d", len(output))
	} else if output["a"] != 101 || output["b"] != 37 {
		t.Errorf("Expected output 101 and 37. Got: %d and %d", output["a"], output["b"])
	}
}

func TestEnqueue(t *testing.T) {
	testPQueue, err := NewPQueue("./", 10)
	if err != nil {
		t.Fatal(err)
	}
	defer testPQueue.Close()

	// Enqueue 50 messages
	for p := one; p <= five; p++ {
		for n := 1; n <= 10; n++ {
			err := testPQueue.Enqueue(p, NewMessagef("test message %d-%d", p, n))
			if err != nil {
				t.Error(err)
			}
		}
	}

	for p := one; p <= five; p++ {
		s, err := testPQueue.Size(p)
		if err != nil {
			t.Error(err)
		} else if s != 10 {
			t.Errorf("Expected queue size 10 for priority %d. Got: %d", p, s)
		}
	}
}

func TestDequeueDeep(t *testing.T) {
	rng := uint(2)

	testPQueue, err := NewPQueue("./", rng)
	if err != nil {
		t.Fatal(err)
	}
	defer testPQueue.Close()

	//Put them in in reverse priority order
	for p := one; p < rng; p++ {
		for n := 1; n <= 50; n++ {
			err := testPQueue.Enqueue(p, NewMessagef("test message %d-%d", p, n))
			if err != nil {
				t.Error(err)
			}
		}
	}

	for p := uint(rng) - 1; p >= one; p-- {
		for n := 1; n <= 50; n++ {
			mStrComp := fmt.Sprintf("test message %d-%d", p, n)
			m, err := testPQueue.Dequeue()
			if err != nil {
				t.Error("Error dequeueing:", err)
			}
			mStr := m.String()
			if mStr != mStrComp {
				t.Errorf("Expected message: \"%s\" got: \"%s\"", mStrComp, mStr)
			}
			if m.Priority() != p {
				t.Errorf("Expected priority: %d, got: %d", p, m.Priority())
			}
		}
	}

	for p := one; p < rng; p++ {
		s, err := testPQueue.Size(p)
		if err != nil {
			t.Error(err)
		} else if s != 0 {
			t.Errorf("Expected queue size 0 for priority %d. Got: %d", p, s)
		}
	}
}

func TestDequeueWide(t *testing.T) {
	rng := uint(260) // more than 256

	testPQueue, err := NewPQueue("./", rng)
	if err != nil {
		t.Fatal(err)
	}
	defer testPQueue.Close()

	//Put them in in reverse priority order
	for p := one; p < rng; p++ {
		err := testPQueue.Enqueue(p, NewMessagef("test message %d", p))
		if err != nil {
			t.Error(err)
		}
	}

	for p := uint(rng) - 1; p >= one; p-- {
		mStrComp := fmt.Sprintf("test message %d", p)
		m, err := testPQueue.Dequeue()
		if err != nil {
			t.Error("Error dequeueing:", err)
		}
		mStr := m.String()
		if mStr != mStrComp {
			t.Errorf("Expected message: \"%s\" got: \"%s\"", mStrComp, mStr)
		}
		if m.Priority() != p {
			t.Errorf("Expected priority: %d, got: %d", p, m.Priority())
		}
	}

	for p := one; p < rng; p++ {
		s, err := testPQueue.Size(p)
		if err != nil {
			t.Error(err)
		} else if s != 0 {
			t.Errorf("Expected queue size 0 for priority %d. Got: %d", p, s)
		}
	}
}

func TestRequeue(t *testing.T) {
	testPQueue, err := NewPQueue("./", 256)
	if err != nil {
		t.Fatal(err)
	}
	defer testPQueue.Close()

	for p := five; p >= one; p-- {
		err := testPQueue.Enqueue(p, NewMessagef("test message %d", p))
		if err != nil {
			t.Error(err)
		}
	}

	mp5, err := testPQueue.Dequeue()
	if err != nil {
		t.Error(err)
	} else if mp5.String() != "test message 5" {
		t.Errorf("Expected: \"%s\", got: \"%s\"", "test message 5", mp5.String())
	}

	//Remove the priority 4 message
	mp4, err := testPQueue.DequeueString()
	if err != nil {
		t.Error(err)
	} else if mp4 != "test message 4" {
		t.Errorf("Expected: \"%s\", got: \"%s\"", "test message 4", mp4)
	}

	//Re-enqueue the message at priority 5
	err = testPQueue.Requeue(5, mp5)
	if err != nil {
		t.Error(err)
	}

	// and it should be the first to emerge
	mp5, err = testPQueue.Dequeue()
	if err != nil {
		t.Error(err)
	} else if mp5.String() != "test message 5" {
		t.Errorf("Expected: \"%s\", got: \"%s\"", "test message 5", mp5.String())
	}
}

func TestGoroutines(t *testing.T) {
	testPQueue, err := NewPQueue("./", 256)
	if err != nil {
		t.Fatal(err)
	}
	defer testPQueue.Close()

	var wg sync.WaitGroup

	if testPQueue.ApproxSize() != 0 {
		t.Errorf("Expected total size 0. Got: %d", testPQueue.ApproxSize())
	}

	for g := 1; g <= 5; g++ {
		wg.Add(1)
		go func() {
			rand.Seed(time.Now().Unix())
			time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
			for p := one; p <= 5; p++ {
				for n := one; n <= 2; n++ {
					err := testPQueue.Enqueue(p, NewMessagef("test message %d", p))
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()

	if testPQueue.ApproxSize() != 50 {
		t.Errorf("Expected total size 50. Got: %d", testPQueue.ApproxSize())
	}

	for p := one; p <= five; p++ {
		s, err := testPQueue.Size(p)
		if err != nil {
			t.Error(err)
		}
		if s != 10 {
			t.Errorf("Expected queue size 10 for priority %d. Got: %d", p, s)
		}
	}

	for p := five; p >= one; p-- {
		for n := 1; n <= 10; n++ {
			mStrComp := fmt.Sprintf("test message %d", p)
			m, err := testPQueue.Dequeue()
			if err != nil {
				t.Error("Error dequeueing:", err)
			}
			mStr := m.String()
			if mStr != mStrComp {
				t.Errorf("Expected message: \"%s\" got: \"%s\"", mStrComp, mStr)
			}
			if m.Priority() != p {
				t.Errorf("Expected priority: %d, got: %d", p, m.Priority())
			}
		}

		if testPQueue.ApproxSize() != int64(p-1)*10 {
			t.Errorf("Expected total size %d. Got: %d", (p-1)*10, testPQueue.ApproxSize())
		}
	}

	for p := one; p <= five; p++ {
		s, err := testPQueue.Size(p)
		if err != nil {
			t.Error(err)
		}
		if s != 0 {
			t.Errorf("Expected queue size 0 for priority %d. Got: %d", p, s)
		}
	}
}

func TestRetainOnClose(t *testing.T) {
	testPQueue, err := NewPQueue("testRetain.db", 256)
	if err != nil {
		t.Fatal(err)
	}
	testPQueue.RetainOnClose = true
	testPQueue.Close()

	err = os.Remove("testRetain.db")
	if err != nil {
		t.Error(err)
	}
}

func benchmarkPQueue(b *testing.B, rng uint) {
	queue, err := NewPQueue("./", rng)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		for p := zero; p < rng; p++ {
			queue.EnqueueString(p, "test message")
		}

		for p := zero; p < rng; p++ {
			_, err := queue.Dequeue()
			if err != nil {
				b.Error(err)
			}
		}
	}

	err = queue.Close()
	if err != nil {
		b.Error(err)
	}
}

func BenchmarkPQueue1(b *testing.B) {
	benchmarkPQueue(b, 1)
}

func BenchmarkPQueue10(b *testing.B) {
	benchmarkPQueue(b, 10)
}

func BenchmarkPQueue100(b *testing.B) {
	benchmarkPQueue(b, 100)
}

func BenchmarkPQueue1000(b *testing.B) {
	benchmarkPQueue(b, 1000)
}
