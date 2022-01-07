package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/acentior/chat-app/internal"
	"github.com/heroku/x/hredis/redigo"
	"github.com/sirupsen/logrus"
)

var (
	waitTimeout                      = time.Second * 10
	log                              = logrus.WithField("cmd", "chat-app")
	waitingMessage, availableMessage []byte
	waitSleep                        = time.Second * 10
)

func init() {
	var err error
	waitingMessage, err = json.Marshal(internal.Message{
		User: "system",
		Data: []byte("Waiting for redis to be available. Messaging won't work until redis is available"),
	})
	if err != nil {
		panic(err)
	}
	availableMessage, err = json.Marshal(internal.Message{
		User: "system",
		Data: []byte("Redis is now available & messaging is now possible"),
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	redisURL := "redis://localhost:6379"
	redisPool, err := redigo.NewRedisPoolFromURL(redisURL)
	if err != nil {
		log.WithField("url", redisURL).Fatal("Unable to create Redis pool")
	}

	rr := internal.NewRedisReceiver(redisPool)
	rw := internal.NewRedisWriter(redisPool)

	go func() {
		for {
			waited, err := redigo.WaitForAvailability(redisURL, waitTimeout, rr.Wait)
			if !waited || err != nil {
				log.WithFields(logrus.Fields{"waitTimeout": waitTimeout, "err": err}).Fatal("Redis not available by timeout!")
			}
			rr.Broadcast(availableMessage)
			err = rr.Run()
			fmt.Println("till here")
			if err == nil {
				break
			}
			log.Error(err)
		}
	}()

	go func() {
		for {
			waited, err := redigo.WaitForAvailability(redisURL, waitTimeout, nil)
			if !waited || err != nil {
				log.WithFields(logrus.Fields{"waitTimeout": waitTimeout, "err": err}).Fatal("Redis not available by timeout!")
			}
			err = rw.Run()
			if err == nil {
				break
			}
			log.Error(err)
		}
	}()
	fmt.Println("Startin the server")
	internal.StartServer(redisPool, rr, rw)
}
