package callp

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"time"
)

func exitOnErr(msg string, err error) {
	if err != nil {
		log.Println(msg, err)
		os.Exit(1)
	}
}

// Pricer is the command which will enable pricing and communicatino between two instances.
func Pricer(cmdPath string, write, read chan string, close chan bool) {
	cmd := exec.Command(cmdPath)
	in, _ := cmd.StdinPipe()
	out, _ := cmd.StdoutPipe()
	buf := bufio.NewReader(out)
	err := cmd.Start()
	exitOnErr("start", err)

	running := true
	go func() {
		for running {
			data, _ := buf.ReadString('\n')
			read <- data
		}
	}()

loop:
	for {
		select {
		case <-close:
			running = false
			in.Close()
			break loop
		case msg := <-write:
			in.Write([]byte(msg + "\n"))
		case <-time.After(time.Second * 2):
			log.Println("No activity for 120 seconds")
			break loop
		}
	}
}
