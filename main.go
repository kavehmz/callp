package main

import (
	"time"

	"github.com/kavehmz/callp/callp"
)

func main() {
	quit := make(chan bool)
	go callp.Plan(quit)

	time.Sleep(time.Second * 15)
	quit <- true

}
