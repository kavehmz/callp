package callp

import (
	"os"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	os.Setenv("REDIS_URL", "")
	p := newPool("")
	_, err := p.Dial()
	if err == nil {
		t.Error("Connection didn't fail with wrong REDIS_URL")
	}

	p = newPool("redis://127.0.0.1:6379/0")
	c := p.Get()
	c.Close()
	c = p.Get()
}

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
