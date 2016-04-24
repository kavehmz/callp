package callp

import (
	"testing"
	"time"
)

func TestSubscribe(t *testing.T) {
	tick := make(chan string)
	psc := subscribe("feed::frxUSDJPY", tick)
	tc := readPool.Get()
	tc.Do("PUBLISH", "feed::frxUSDJPY", "test")

	select {
	case data := <-tick:
		if data != "test" {
			t.Error("Message incorrect")
		}
	case <-time.After(time.Second):
		t.Error("No message received")
	}

	go psc.Close()

	select {
	case <-tick:
	case <-time.After(time.Second):
		t.Error("Connection still open")
	}
}
