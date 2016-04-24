package main

import (
	"fmt"
	"time"

	"github.com/kavehmz/callp/callp"
)

func main() {
	tick := make(chan string, 1)
	psc := callp.Subscribe("FEED::R_25", tick)
	fmt.Println(<-tick)
	fmt.Println("Done00")
	go psc.Close()
	fmt.Println("Done0")
	close(tick)
	fmt.Println("Done1")
	<-tick
	fmt.Println("Done")
	select {
	case <-tick:
		fmt.Println("Done")
	default:
		fmt.Println("not Done")
	}

	time.Sleep(time.Second * 100)

	// quit := make(chan bool)
	// go callp.Plan(quit)
	//
	// time.Sleep(time.Second * 15)
	// quit <- true

}
