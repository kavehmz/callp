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
