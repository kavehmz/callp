package callp

import (
	"fmt"
	"log"

	"github.com/garyburd/redigo/redis"
)

func subscriber(redisChannel string, tick chan string) {
	c := readPool.Get()
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

func publish(req pricinigRequest, msg string) {
	fmt.Println(req, msg)
}
