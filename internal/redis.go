package internal

import (
	"fmt"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	channelList = []string{"room1", "room2", "room3"}
)

type redisReceiver struct {
	pool           *redis.Pool
	messages       chan []byte
	newConnections chan *websocket.Conn
	rmConnections  chan *websocket.Conn
}

func NewRedisReceiver(pool *redis.Pool) redisReceiver {
	return redisReceiver{
		pool:           pool,
		messages:       make(chan []byte, 1000), // 1000 is arbitrary
		newConnections: make(chan *websocket.Conn),
		rmConnections:  make(chan *websocket.Conn),
	}
}

func (rr *redisReceiver) Wait(_ time.Time) error {
	rr.Broadcast([]byte("Delivering a waiting message"), roomId)
	return nil
}

func (rr *redisReceiver) Run() error {
	l := log.WithField("channel", channelList)
	conn := rr.pool.Get()
	defer conn.Close()
	psc := redis.PubSubConn{Conn: conn}

	err := psc.PSubscribe(RoomsList)
	if err != nil {
		errors.Wrap(err, "failed to subscribe to multiple channels")
		return
	}

	go rr.ConnHandler()

	done := make(chan error, 1)

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			if err := rr.onMessage(v.Channel, v.Data); err != nil {
				done <- err
				return 
			}	
		case redis.Subscription:
			switch  v.Count() {
			case len(v.Channels):
					if err = onStart(); err != nil {
						done <- err
						return
					}
			case 0:
				done <- nil
				return
			}
		case redis.Error:
			done <- v
			return
		default:
			l.WithField("v", v).Info("Unknown Command received")
		}
	}
}

func (rr *redisReceiver) onMessage(channel string, data []byte) error {
	for _, v := range RoomsList {
		if channel == v.ChannelName {
			rr.BroadCast(channel, data)
		}
	}
}

func (rr *redisReceiver) Broadcast(msg []byte, channel string) {
	rr.messages <- msg
}

func (rr *redisReceiver) Register(conn *websocket.Conn) {
	rr.newConnections <- conn
}

func (rr *redisReceiver) DeRegister(conn *websocket.Conn) {
	rr.rmConnections <- conn
}

func (rr *redisReceiver) ConnHandler() {
	conns := make([]*websocket.Conn, 0)
	var channelName string
	for {
		select {
		case msg := <-rr.messages:
			if _, v := channelList {
			}
		case conn := <-rr.newConnections:
			conns = append(conns, conn)
		case conn := <-rr.rmConnections:
			conns = RemoveConn(conns, conn)
		}
	}
}

func RemoveConn(conns []*websocket.Conn, remove *websocket.Conn) []*websocket.Conn {
	var i int
	var found bool
	for i = 0; i < len(conns); i++ {
		if conns[i] == remove {
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("conns: %#v\nconn: %#v\n", conns, remove)
		panic("Conn not found")
	}
	copy(conns[i:], conns[i+1:]) // shift down
	conns[len(conns)-1] = nil    // nil last element
	return conns[:len(conns)-1]  // truncate slice
}

type redisWriter struct {
	pool     *redis.Pool
	messages chan []byte
}

func NewRedisWriter(pool *redis.Pool) redisWriter {
	return redisWriter{
		pool:     pool,
		messages: make(chan []byte, 10000),
	}
}

func (rw *redisWriter) Run() error {
	conn := rw.pool.Get()
	defer conn.Close()
	for data := range rw.messages {
		if err := writeToRedis(conn, data, channel); err != nil {
			rw.Publish(data) // attempt to redeliver later
			return err
		}
	}
	return nil
}

func writeToRedis(conn redis.Conn, data []byte, roomId string) error {
	if err := conn.Send("PUBLISH", channelList, data); err != nil {
		return errors.Wrap(err, "Unable to publish message to Redis")
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrap(err, "Unable to flush published message to Redis")
	}
	return nil
}

// publish to Redis via channel.
func (rw *redisWriter) Publish(data []byte, roomID string) {
	rw.messages <- data
}
