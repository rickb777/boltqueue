package boltqueue

import (
	"fmt"
	"sync"
	"testing"
)

func TestIChanUsingSend(t *testing.T) {
	ich, err := NewIChan("./")
	if err != nil {
		t.Fatal(err)
	}

	ich.SetErrorHandler(func(e error) {
		panic(e)
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		i := 1
		c := ich.ReceiveEnd()
		var v []byte
		ok := true
		for ok {
			v, ok = <-c
			if ok {
				s := string(v)
				expected := fmt.Sprintf("%d", i)
				if s != expected {
					t.Errorf("Expected: \"%s\", got: \"%s\"", expected, s)
				}
				i++
			}
		}
		wg.Done()
	}()

	for p := one; p <= 100; p++ {
		err = ich.SendString(fmt.Sprintf("%d", p))
		if err != nil {
			t.Fatal(err)
		}
	}
	ich.Close()

	wg.Wait()
}

func TestIChanUsingChannel(t *testing.T) {
	ich, err := NewIChan("./")
	if err != nil {
		t.Fatal(err)
	}

	ich.SetErrorHandler(func(e error) {
		panic(e)
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		i := 1
		c := ich.ReceiveEnd()
		var v []byte
		ok := true
		for ok {
			v, ok = <-c
			if ok {
				s := string(v)
				expected := fmt.Sprintf("%d", i)
				if s != expected {
					t.Errorf("Expected: \"%s\", got: \"%s\"", expected, s)
				}
				i++
			}
		}
		wg.Done()
	}()

	for p := one; p <= 100; p++ {
		v := []byte(fmt.Sprintf("%d", p))
		ich.SendEnd() <- v
	}
	close(ich.SendEnd())

	wg.Wait()
}
