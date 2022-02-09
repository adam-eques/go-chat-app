package internal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	log                              = logrus.WithField("cmd", "chat-app")
	waitingMessage, availableMessage []byte
	waitSleep                        = time.Second * 10
)

func init() {
	var err error
	waitingMessage, err = json.Marshal(Message{
		User: "system",
		Data: []byte("Waiting for redis to be available. Messaging won't work until redis is available"),
	})
	if err != nil {
		panic(err)
	}
	availableMessage, err = json.Marshal(Message{
		User: "system",
		Data: []byte("Redis is now available & messaging is now possible"),
	})
	if err != nil {
		panic(err)
	}
}

type redisReceiver struct {
	pool           *redis.Pool
	roomID         string
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
	rr.Broadcast(waitingMessage)
	time.Sleep(waitSleep)
	return nil
}

func (rr *redisReceiver) Run(roomID string) error {
	l := log.WithField("channel", rr.roomID)
	conn := rr.pool.Get()
	defer conn.Close()
	psc := redis.PubSubConn{Conn: conn}
	psc.Subscribe(roomID)
	go rr.ConnHandler()
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			l.WithField("message", string(v.Data)).Info("Redis Message Received")
			rr.Broadcast(v.Data)
		case redis.Subscription:
			l.WithFields(logrus.Fields{
				"kind":   v.Kind,
				"count":  v.Count,
				"roomID": roomID,
			}).Println("Redis Subscription Received to %v", roomID)
		case error:
			return errors.Wrap(v, "Error while subscribed to Redis channel")
		default:
			l.WithField("v", v).Info("Unknown Redis receive during subscription")
		}
	}
}

func (rr *redisReceiver) Broadcast(msg []byte) {
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
	for {
		select {
		case msg := <-rr.messages:
			for _, conn := range conns {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					log.WithFields(logrus.Fields{
						"data": msg,
						"err":  err,
						"conn": conn,
					}).Error("Error writting data to connection! Closing and removing Connection")
					conns = RemoveConn(conns, conn)
				}
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

func (rw *redisWriter) Run(roomId string) error {
	conn := rw.pool.Get()
	defer conn.Close()
	for data := range rw.messages {
		if err := writeToRedis(conn, data, roomId); err != nil {
			rw.Publish(data) // attempt to redeliver later
			return err
		}
	}
	return nil
}

func writeToRedis(conn redis.Conn, data []byte, roomId string) error {
	if err := conn.Send("PUBLISH", roomId, data); err != nil {
		return errors.Wrap(err, "Unable to publish message to Redis")
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrap(err, "Unable to flush published message to Redis")
	}
	return nil
}

// publish to Redis via channel.
func (rw *redisWriter) Publish(data []byte) {
	rw.messages <- data
}
