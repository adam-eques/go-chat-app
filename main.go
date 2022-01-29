package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/heroku/x/hredis/redigo"
	"github.com/joho/godotenv"
	"github.com/midepeter/chat-app/internal"
	"github.com/sirupsen/logrus"
)

var (
	waitTimeout                      = time.Minute * 1
	log                              = logrus.WithField("cmd", "chat-app")
	waitingMessage, availableMessage []byte
	waitSleep                        = time.Minute * 1
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
	err := godotenv.Load()
	if err != nil {
		errors.New("couldnt load environment variables")
	}
	redisHost := os.Getenv("REDISHOST")
	redisPort := os.Getenv("REDISPORT")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	redisURL := fmt.Sprintf("redis://%s:%s", redisHost, redisPort)
	fmt.Println(redisURL)
	const maxconnections = 10
	redisPool := &redis.Pool{
		MaxIdle: maxconnections,
		Dial:    func() (redis.Conn, error) { return redis.Dial("tcp", redisAddr) },
	}
	//	redisPool, err := redigo.NewRedisPoolFromURL(redisURL)
	if err != nil {
		fmt.Println("Connectionpool returned successufull")
		log.WithField("url", redisURL).Fatal("Unable to create Redis pool")
	}

	rr := internal.NewRedisReceiver(redisPool)
	rw := internal.NewRedisWriter(redisPool)

	go func() {
		for {
			waited, err := redigo.WaitForAvailability(redisURL, waitTimeout, rr.Wait)
			if !waited || err != nil {
				fmt.Println(err)
				log.WithFields(logrus.Fields{"waitTimeout": waitTimeout, "err": err}).Fatal("Redis not available by timeout!")
			}
			rr.Broadcast(availableMessage)
			err = rr.Run()
			fmt.Println("till here")
			if err == nil {
				break
			}
			log.Error(err)
			fmt.Println("finally connected")
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
	fmt.Println("Starting the server")
	internal.StartServer(redisPool, rr, rw)
}
