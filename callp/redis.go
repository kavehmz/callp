package callp

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

func newPool(redisURL string) *redis.Pool {
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
	redisRead := os.Getenv("REDISREAD_URL")
	redisWrite := os.Getenv("REDISWRITE_URL")
	if os.Getenv("REDIS_URL") != "" {
		redisRead = os.Getenv("REDIS_URL")
		redisWrite = os.Getenv("REDIS_URL")
	}
	readPool = newPool(redisRead)
	writePool = newPool(redisWrite)
}

func subscribe(redisChannel string, tick chan string) redis.PubSubConn {
	c := readPool.Get()

	psc := redis.PubSubConn{Conn: c}
	psc.Subscribe(redisChannel)
	go func() {
	loop:
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				tick <- string(v.Data)
			case error:
				break loop
			}
		}
		close(tick)
	}()
	return psc
}

func workStillValid(id int64) bool {
	c := readPool.Get()
	defer c.Close()
	ttl, _ := redis.Int64(c.Do("TTL", "work::"+strconv.FormatInt(id, 10)))
	if ttl > 0 {
		return true
	}
	return false
}

func reqByID(id int64) (req PricinigRequest) {
	c := readPool.Get()
	defer c.Close()
	msg, _ := redis.String(c.Do("GET", "work::"+strconv.FormatInt(id, 10)))
	json.Unmarshal([]byte(msg), &req)
	req.ID = id
	return req
}

func nextJob(nextJob chan PricinigRequest, quit chan bool) {
	c := readPool.Get()
	defer c.Close()
loop:
	for {
		jobID, _ := redis.Int64(c.Do("INCR", "work::provide"))
		for {
			select {
			case <-quit:
				break loop
			default:
				break
			}
			lastestRequest, _ := redis.Int64(c.Do("GET", "work::offer"))
			if lastestRequest >= jobID {
				nextJob <- reqByID(jobID)
				break
			} else {
				time.Sleep(time.Duration(TimeoutMultiplier * WaitIfNoJob))
			}
		}
	}
}

func publish(redisChannel string, pub chan Read, quit chan bool) {
	c := readPool.Get()

loop:
	for {
		select {
		case t := <-pub:
			c.Do("PUBLISH", redisChannel, t.data)
		case <-quit:
			break loop
		}
	}
	go c.Close()
}
