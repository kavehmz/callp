package callp

import (
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
	read := make(chan Read, 1)
	err := make(chan error, 1)
	pricerQuit := make(chan bool, 1)
	go Pricer(PricingScript, write, read, pricerQuit, err)

	publisherQuit := make(chan bool, 1)
	go publish(req.Trigger, read, publisherQuit)

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
			publisherQuit <- true
			break loop
		case e := <-err:
			log.Println(e)
			pricerQuit <- true
			break loop
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

	nextJobID := make(chan PricinigRequest)
	quitNextJob := make(chan bool)
	go nextJob(nextJobID, quitNextJob)

	worker := make(chan bool, 60)
loop:
	for {
		select {
		case req := <-nextJobID:
			q := make(chan bool)
			go streamControl(req, q)
			go streamPrice(worker, req, q)
			worker <- true
		case <-quit:
			break loop
		}
	}
}
