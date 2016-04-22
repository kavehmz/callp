package callp

import (
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
