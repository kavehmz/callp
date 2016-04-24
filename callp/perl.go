package callp

import (
	"bufio"
	"errors"
	"log"
	"os/exec"
	"time"
)

// Pricer is an interactive command-line executer. It expects the script to communicate over STDIN and STDOUT.
// read,write and close channels are means of communicating with the script.
// Pricer will exit if the script exits.
func Pricer(cmdPath string, write, read chan string, close chan bool, err chan error) {
	cmd := exec.Command(cmdPath)
	in, _ := cmd.StdinPipe()
	out, _ := cmd.StdoutPipe()
	buf := bufio.NewReader(out)
	e := cmd.Start()
	if e != nil {
		err <- e
		return
	}

	readTimeout := time.NewTimer(time.Duration(0))
	readTimeout.Stop()
	running := true
	var t time.Time
	go func() {
		for running {
			data, _ := buf.ReadString('\n')
			log.Println("Pricing time", data, time.Now().Sub(t).Nanoseconds())
			readTimeout.Stop()
			read <- data
		}
	}()

	appEnded := make(chan bool)
	go func() {
		e := cmd.Wait()
		if e != nil {
			err <- e
		}
		appEnded <- true
	}()

loop:
	for {
		select {
		case msg := <-write:
			readTimeout.Reset(time.Duration(PricerReadTimeout * TimeoutMultiplier))
			t = time.Now()
			in.Write([]byte(msg + "\n"))
		case <-appEnded:
			running = false
			in.Close()
			break loop
		case <-close:
			running = false
			in.Close()
			break loop
		case <-readTimeout.C:
			err <- errors.New("Read timeout")
			running = false
			in.Close()
			break loop
		}
	}
}
