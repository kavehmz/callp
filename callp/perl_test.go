package callp

import (
	"testing"
	"time"
)

func TestPricerWrongPath(t *testing.T) {
	write := make(chan string, 2)
	read := make(chan Read, 2)
	quit := make(chan bool, 1)
	err := make(chan error, 1)
	go Pricer("../dummy.pl", write, read, quit, err)
	select {
	case <-err:
	case <-time.After(time.Second):
		t.Error("No error when path was wrong")
	}
}

func TestPricerReadMustTimeout(t *testing.T) {
	write := make(chan string, 2)
	read := make(chan Read, 2)
	quit := make(chan bool, 1)
	err := make(chan error, 1)
	go Pricer("../pricer.pl", write, read, quit, err)
	write <- "EN"
	write <- "Data"

	write <- "1500000"
	select {
	case <-err:
	case <-read:
		t.Error("Read returned data when it was supposed to timeout")
	case <-time.After(time.Duration(PricerReadTimeout * TimeoutMultiplier * 2)):
		t.Error("Read did not timeout")
	}
}

func TestPricerEarlyEndWillCauseError(t *testing.T) {
	write := make(chan string, 2)
	read := make(chan Read, 2)
	quit := make(chan bool, 1)
	err := make(chan error, 1)
	go Pricer("../pricer.pl", write, read, quit, err)
	write <- "EN"
	write <- "1"

	write <- "1"
	select {
	case <-err:
	case <-time.After(time.Second):
		t.Error("No error after script exited by its own")
	}
}

func TestPricerCloseWillEnd(t *testing.T) {
	write := make(chan string, 2)
	read := make(chan Read, 2)
	quit := make(chan bool, 1)
	err := make(chan error, 1)

	stopped := make(chan bool)
	go func() {
		Pricer("../pricer.pl", write, read, quit, err)
		stopped <- true
	}()
	quit <- true
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Error("Process didn't quit")
	}
}

func TestPricerNormalPricingWithDuration(t *testing.T) {
	write := make(chan string, 2)
	read := make(chan Read, 2)
	quit := make(chan bool, 1)
	err := make(chan error, 1)
	go Pricer("../pricer.pl", write, read, quit, err)
	write <- "EN"
	write <- "1"

	write <- "10000" //timeout
	select {
	case msg := <-read:
		if msg.duration.Nanoseconds() < 10000000 || msg.data == "" {
			t.Error("Wrong data")
		}
	case <-time.After(time.Duration(PricerReadTimeout * TimeoutMultiplier * 2)):
		t.Error("Read did not reply")
	}
}
