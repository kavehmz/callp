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

func TestPublishSubscribe(t *testing.T) {
	tick := make(chan string)
	psc := subscribe("feed::frxUSDJPY", tick)

	quit := make(chan bool, 1)
	read := make(chan Read)
	go publish("feed::frxUSDJPY", read, quit)
	read <- Read{data: "test"}
	quit <- true

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

func TestNextJob(t *testing.T) {
	c := readPool.Get()
	defer c.Close()
	c.Do("SET", "work::provide", "0")
	c.Do("SET", "work::offer", "2")

	c.Do("SET", "work::1", `{"id": 1, "lang": "FR", "params": "test_params", "md5": "abcdefghijhklmnopqrstuvwxyz", "trigger": "R_25"}`)
	c.Do("SET", "work::2", `{"id": 2, "lang": "EN", "params": "test_params", "md5": "abcdefghijhklmnopqrstuvwxyz", "trigger": "R_25"}`)

	next := make(chan PricinigRequest)
	quit := make(chan bool)
	go nextJob(next, quit)

	select {
	case req := <-next:
		if req.ID != 1 {
			t.Error("Wrong job number")
		}
	case <-time.After(time.Second):
		t.Error("No job found")
	}

	select {
	case req := <-next:
		if req.ID != 2 {
			t.Error("Wrong job number")
		}
	case <-time.After(time.Second):
		t.Error("No job found")
	}

	select {
	case <-next:
		t.Error("There was a next job without any job provider!")
	case <-time.After(time.Duration(TimeoutMultiplier * WaitIfNoJob)):
		break
	}
	quit <- true
}
