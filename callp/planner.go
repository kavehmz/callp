package callp

import (
	"log"
	"time"
)

const done = 0

// PricinigRequest will store the pricing details set by job creators.
type PricinigRequest struct {
	// Unique job id
	ID int64 `json:"id"`
	// Languge used in messages send by pricer.
	Lang string `json:"lang"`
	// Params will store pricing details. Perl script must be able to deserealize this.
	Params string `json:"params"`
	// MD5 is unique idenfitfier of pricing. This will be used to avoid having to worker generating the same price.
	MD5 string `json:"md5"`
	// Trigger is a redis channel that will trigger the next pricing.
	Trigger string `json:"trigger"`
}

func streamPrice(worker chan bool, job chan int64, req PricinigRequest) {
	write := make(chan string, 3)
	read := make(chan Read, 1)
	err := make(chan error, 1)
	quit := make(chan bool, 1)
	go Pricer(PricingScript, write, read, quit, err)

	publisherQuit := make(chan bool, 1)
	go publish(req.Trigger, read, publisherQuit)

	tick := make(chan string)
	psc := subscribe(req.Trigger, tick)

	write <- req.Lang
	write <- req.Params
loop:
	for {
		select {
		case e := <-err:
			log.Println(e)
			quit <- true
			publisherQuit <- true
			break loop
		case signal := <-tick:
			write <- signal
			if !workStillValid(req.ID) {
				quit <- true
				publisherQuit <- true
				break loop
			}
		case <-time.After(time.Duration(PricerInactivityTimeout * TimeoutMultiplier)):
			log.Println("No activity in price streamer for more than ", time.Duration(PricerInactivityTimeout*TimeoutMultiplier).Seconds(), "second.")
			quit <- true
			publisherQuit <- true
			break loop
		}
	}
	go psc.Close()
	<-worker
	job <- done
}

// Plan select the next job, send it to a price streamer and notify that job has started
func Plan(quit chan bool, job chan int64) {
	next := make(chan PricinigRequest)
	quitNext := make(chan bool)
	go nextJob(next, quitNext)

	worker := make(chan bool, MaxConcurrentWorkers)
loop:
	for {
		select {
		case req := <-next:
			go streamPrice(worker, job, req)
			job <- req.ID
			worker <- true
		case <-quit:
			break loop
		}
	}
}
