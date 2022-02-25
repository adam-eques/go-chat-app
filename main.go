package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/acentior/chat-app/internal"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("cmd", "chat-app")

func main() {
	err := godotenv.Load()
	if err != nil {
		errors.New("couldnt load environment variables")
	}
	redisHost := os.Getenv("REDISHOST")
	redisPort := os.Getenv("REDISPORT")
	redisURL := fmt.Sprintf("%s:%s", redisHost, redisPort)

	const maxconnections = 10
	redisPool := &redis.Pool{
		MaxIdle: maxconnections,
		Dial:    func() (redis.Conn, error) { return redis.Dial("tcp", redisURL) },
	}

	rr := internal.NewRedisReceiver(redisPool)

	rw := internal.NewRedisWriter(redisPool)
	go rr.Run()

	go rw.Run()

	fmt.Println("Starting the pp server")
	internal.StartServer(redisPool, rr, rw)
}
