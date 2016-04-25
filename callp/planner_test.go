package callp

import (
	"testing"
	"time"
)

func TestPlanner(t *testing.T) {
	c := readPool.Get()
	defer c.Close()
	c.Do("SET", "work::1", `{"lang": "FR", "params": "test_params", "md5": "abcdefghijhklmnopqrstuvwxyz", "trigger": "R_25"}`)
	c.Do("SET", "work::2", `{"lang": "EN", "params": "test_params", "md5": "abcdefghijhklmnopqrstuvwxyz", "trigger": "R_25"}`)

	c.Do("SET", "work::provide", "0")
	c.Do("SET", "work::offer", "1")

	quit := make(chan bool)
	job := make(chan int64, MaxConcurrentWorkers)
	go Plan(quit, job)

	select {
	case <-job:
	case <-time.After(time.Second):
		t.Error("No job started")
	}
	quit <- true

}
