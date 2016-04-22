package callp

import (
	"fmt"
	"log"
	"os"
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
				return nil, err
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
	readPool = newReadPool(os.Getenv("REDISREAD_URL"))
	writePool = newReadPool(os.Getenv("REDISWRITE_URL"))
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

func nextJobID(prefix string) int64 {
	c := readPool.Get()
	defer c.Close()

	jobID, _ := redis.Int64(c.Do("INCR", "work::offer"))
	for {
		lastestRequest, _ := redis.Int64(c.Do("GET", "work::request"))
		if lastestRequest >= jobID {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	return jobID
}

func publish(req pricinigRequest, msg string) {
	fmt.Println(req, msg)
}
