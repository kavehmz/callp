package callp

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

func newReadPool(redisURL string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(redisURL)
			if err != nil {
				log.Fatalln(err)
			}
			_, err = c.Do("PING")
			if err != nil {
				log.Fatalln("rr", err)
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

var readPool, writePool *redis.Pool

func init() {
	redisRead := os.Getenv("REDISREAD_URL")
	redisWrite := os.Getenv("REDISWRITE_URL")
	if os.Getenv("REDIS_URL") != "" {
		redisRead = os.Getenv("REDIS_URL")
		redisWrite = os.Getenv("REDIS_URL")
	}
	readPool = newReadPool(redisRead)
	writePool = newReadPool(redisWrite)
}

func subscriber(redisChannel string, tick chan string) {
	c := readPool.Get()
	defer c.Close()

	psc := redis.PubSubConn{Conn: c}
	psc.Subscribe(redisChannel)
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			tick <- string(v.Data)
		case error:
			log.Println("Subscriber fail in the middle of listening")
			return
		}
	}
}

func reqByID(id int64) (req PricinigRequest) {
	c := readPool.Get()
	defer c.Close()

	msg, _ := redis.String(c.Do("GET", "work::"+strconv.FormatInt(id, 10)))
	json.Unmarshal([]byte(msg), &req)
	return PricinigRequest{ID: id, Lang: "FR", Params: "test_params", MD5: time.Now().Format("UnixDate"), Trigger: "R_25"}
	//return req
}

func nextJob(nextJob chan int64) {
	c := readPool.Get()
	defer c.Close()

	for {
		jobID, _ := redis.Int64(c.Do("INCR", "work::provide"))
		for {
			lastestRequest, _ := redis.Int64(c.Do("GET", "work::offer"))
			if lastestRequest >= jobID {
				nextJob <- jobID
				break
			} else {
				fmt.Println("nextJob not there yet", lastestRequest, jobID)
				time.Sleep(time.Millisecond * 1000)
			}
		}
	}
}

func publish(req PricinigRequest, msg string) {
	fmt.Println(req, msg)
}
