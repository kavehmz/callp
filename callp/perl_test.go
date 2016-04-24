package callp

import (
	"testing"
	"time"
)

func TestPricerWrongPath(t *testing.T) {
	write := make(chan string, 2)
	read := make(chan string, 2)
	pricerQuit := make(chan bool, 1)
	err := make(chan error, 1)
	go Pricer("../dummy.pl", write, read, pricerQuit, err)
	select {
	case <-err:
	case <-time.After(time.Second):
		t.Error("No error when path was wrong")
	}
}

func TestPricerReadMustTimeout(t *testing.T) {

	write := make(chan string, 2)
	read := make(chan string, 2)
	pricerQuit := make(chan bool, 1)
	err := make(chan error, 1)
	go Pricer("../pricer.pl", write, read, pricerQuit, err)
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
	read := make(chan string, 2)
	pricerQuit := make(chan bool, 1)
	err := make(chan error, 1)
	go Pricer("../pricer.pl", write, read, pricerQuit, err)
	write <- "EN"
	write <- "1"

	write <- "1500000"
	select {
	case <-err:
	case <-read:
		t.Error("Read returned data when it was supposed to timeout")
	case <-time.After(time.Duration(PricerReadTimeout * TimeoutMultiplier * 2)):
		t.Error("Read did not timeout")
	}
}
