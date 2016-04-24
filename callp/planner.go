package callp

import (
	"fmt"
	"log"
	"time"
)

// PricinigRequest will store the pricing job details
type PricinigRequest struct {
	ID      int64  `json:"id"`
	Lang    string `json:"lang"`
	Params  string `json:"params"`
	MD5     string `json:"md5"`
	Trigger string `json:"trigger"`
}

func streamControl(req PricinigRequest, quit chan bool) {
	time.Sleep(time.Second * 10)
	quit <- true
}

func streamPrice(worker chan bool, req PricinigRequest, quit chan bool) {
	write := make(chan string, 1)
	read := make(chan string, 1)
	err := make(chan error, 1)
	pricerQuit := make(chan bool, 1)
	go Pricer(PricingScript, write, read, pricerQuit, err)

	tick := make(chan string)
	go func() {
		for {
			tick <- "ok"
			time.Sleep(time.Second)
		}
	}()

	write <- req.Lang
	write <- req.Params
loop:
	for {
		select {
		case <-quit:
			pricerQuit <- true
			break loop
		case e := <-err:
			log.Println(e)
			pricerQuit <- true
			break loop
		case msg := <-read:
			publish(req, msg)
		case signal := <-tick:
			write <- signal
		case <-time.After(time.Duration(PricerInactivityTimeout * TimeoutMultiplier)):
			log.Println("No activity in price streamer for more than ", time.Duration(PricerInactivityTimeout*TimeoutMultiplier).Seconds(), "second.")
			break loop
		}
	}

	<-worker
}

// Plan will do planning
func Plan(quit chan bool) {

	c := readPool.Get()
	defer c.Close()
	c.Do("FLUSHALL")
	c.Do("INCR", "work::offer")

	nextJobID := make(chan int64)
	go nextJob(nextJobID)

	worker := make(chan bool, 60)
	var workerID int64
loop:
	for {
		select {
		case id := <-nextJobID:
			fmt.Println("Job stared", id)
			q := make(chan bool)
			req := reqByID(id)
			workerID++
			req.ID = workerID
			go streamControl(req, q)
			go streamPrice(worker, req, q)
			worker <- true
		case <-quit:
			break loop
		}
	}
}
