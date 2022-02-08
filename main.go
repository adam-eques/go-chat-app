package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/acentior/chat-app/internal"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
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

	// redisURL := fmt.Sprintf("redis://%s:%s", redisHost, redisPort)

	const maxconnections = 10
	redisPool := &redis.Pool{
		MaxIdle: maxconnections,
		Dial:    func() (redis.Conn, error) { return redis.Dial("tcp", redisAddr) },
	}

	fmt.Println("Starting the server")
	internal.StartServer(redisPool)
}
