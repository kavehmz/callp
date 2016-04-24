package callp

import (
	"bufio"
	"errors"
	"os/exec"
	"strings"
	"time"
)

// Read includes the main pricing message and other collected data
type Read struct {
	data     string
	duration time.Duration
}

// Pricer is an interactive command-line executer. It expects the script to communicate over STDIN and STDOUT.
// read,write and close channels are means of communicating with the script.
// Pricer will exit if the script exits.
func Pricer(cmdPath string, write chan string, read chan Read, quit chan bool, err chan error) {
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
			readTimeout.Stop()
			read <- Read{data: strings.TrimRight(data, "\n"), duration: time.Now().Sub(t)}
		}
	}()

	appEnded := make(chan bool)
	go func() {
		cmd.Wait()
		if running {
			appEnded <- true
		}
	}()

loop:
	for {
		select {
		case msg := <-write:
			readTimeout.Reset(time.Duration(PricerReadTimeout * TimeoutMultiplier))
			t = time.Now()
			in.Write([]byte(msg + "\n"))
		case <-appEnded:
			err <- errors.New("App ended without signal")
			running = false
			in.Close()
			break loop
		case <-quit:
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
