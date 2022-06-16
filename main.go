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
		log.Fatal(errors.New("couldnt load environment variables"))
	}

	redisHost := os.Getenv("REDISHOST")
	redisPort := os.Getenv("REDISPORT")
	redisURL := fmt.Sprintf("%s:%s", redisHost, redisPort)

	redisPool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisURL)
		},
	}

	fmt.Println("=====Starting the server=====")
	internal.StartServer(redisPool)
}
