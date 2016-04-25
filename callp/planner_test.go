package callp

import (
	"testing"
	"time"
)

func TestPlanner(t *testing.T) {
	PricingScript = "../pricer.pl"
	c := readPool.Get()
	defer c.Close()
	c.Do("SET", "work::1", `{"lang": "FR", "params": "test_params", "md5": "abcdefghijhklmnopqrstuvwxyz", "trigger": "R_25"}`)

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

func TestStreamPriceWithError(t *testing.T) {
	worker := make(chan bool, 1)
	job := make(chan int64, 1)
	req := PricinigRequest{ID: 1, Lang: "EN", Params: "test_params", MD5: "abcdefghijhklmnopqrstuvwxyz", Trigger: "R_25"}
	worker <- true
	PricingScript = "../whatifwrong.pl"
	go streamPrice(worker, job, req)
	select {
	case <-job:
	case <-time.After(time.Second):
		t.Error("Job didn't end when there was an error")
	}
}

func TestStreamPriceTimeout(t *testing.T) {
	worker := make(chan bool, 1)
	job := make(chan int64, 1)
	req := PricinigRequest{ID: 1, Lang: "EN", Params: "test_params", MD5: "abcdefghijhklmnopqrstuvwxyz", Trigger: "R_25"}
	worker <- true
	tmp := PricerInactivityTimeout
	PricerInactivityTimeout = 1
	PricingScript = "../pricer.pl"
	go streamPrice(worker, job, req)
	select {
	case <-job:
	case <-time.After(time.Second):
		t.Error("Job didn't timeout")
	}
	PricerInactivityTimeout = tmp
}

func TestStreamPriceTick(t *testing.T) {
	worker := make(chan bool, 1)
	job := make(chan int64, 1)
	req := PricinigRequest{ID: 1, Lang: "EN", Params: "test_params", MD5: "abcdefghijhklmnopqrstuvwxyz", Trigger: "R_25"}
	worker <- true
	PricingScript = "../pricer.pl"
	tmp := PricerInactivityTimeout
	PricerInactivityTimeout = 100
	go streamPrice(worker, job, req)
	c := readPool.Get()
	defer c.Close()
	c.Do("PUBLISH", "R_25", "1")
	select {
	case <-job:
	case <-time.After(time.Second):
		t.Error("Job didn't timeout")
	}
	PricerInactivityTimeout = tmp
}
