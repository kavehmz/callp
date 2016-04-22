package callp

import (
	"log"
	"time"
)

type pricinigRequest struct {
	lang    string
	params  string
	md5     string
	trigger string
}

func findNextPricingRequest(next chan pricinigRequest) {
	for {
		jobID := nextJobID()

		next <- pricinigRequest{lang: "EN", params: "test_params", md5: time.Now().Format("UnixDate"), trigger: "R_25"}
		time.Sleep(time.Second * 2)
	}
}

func streamControl(req pricinigRequest, quit chan bool) {
	time.Sleep(time.Second * 10)
	quit <- true
}

func streamPrice(worker chan bool, req pricinigRequest, quit chan bool) {
	write := make(chan string, 2)
	read := make(chan string, 2)
	pricerQuit := make(chan bool, 1)
	go Pricer("./pricer.pl", write, read, pricerQuit)

	tick := make(chan string)
	go func() {
		for {
			tick <- "ok"
			time.Sleep(time.Second)
		}
	}()

	write <- req.lang
	write <- req.params
loop:
	for {
		select {
		case <-quit:
			pricerQuit <- true
			break loop
		case msg := <-read:
			publish(req, msg)
		case signal := <-tick:
			write <- signal
		case <-time.After(time.Second * 10):
			log.Println("No activity for 120 seconds")
			break loop
		}
	}

	<-worker
}

// Plan will do planning
func Plan(quit chan bool) {
	nextReq := make(chan pricinigRequest)
	go findNextPricingRequest(nextReq)

	worker := make(chan bool, 60)
loop:
	for {
		select {
		case req := <-nextReq:
			q := make(chan bool)
			go streamControl(req, q)
			go streamPrice(worker, req, q)
			worker <- true
		case <-quit:
			break loop
		case <-time.After(time.Second * 10):
			log.Println("No activity for 120 seconds")
			break loop
		}
	}
}
