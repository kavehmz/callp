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
	var t time.Time
	quitReading := make(chan bool)
	go func() {
	loop:
		for {
			select {
			case <-quitReading:
				break loop
			default:
				data, _ := buf.ReadString('\n')
				readTimeout.Stop()
				read <- Read{data: strings.TrimRight(data, "\n"), duration: time.Now().Sub(t)}
			}
		}
	}()

	// https://github.com/golang/go/issues/9307
	// After this bug is fixed in 1.6.x we can wait for script end without race issue.
	// race check is important enough that I prefer to disable this check for now.
	// appEnded := make(chan bool)
	// go func() {
	// 	cmd.Wait()
	// 	appEnded <- true
	// }()

loop:
	for {
		select {
		case msg := <-write:
			readTimeout.Reset(time.Duration(PricerReadTimeout * TimeoutMultiplier))
			t = time.Now()
			in.Write([]byte(msg + "\n"))
		// case <-appEnded:
		// 	err <- errors.New("App ended")
		// 	break loop
		case <-quit:
			in.Close()
			break loop
		case <-readTimeout.C:
			err <- errors.New("Read timeout")
			in.Close()
			break loop
		}
	}
}
