package internal

import (
	"fmt"
	"io"

	"github.com/acentior/chat-app/firebase"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/gomodule/redigo/redis"
)

func StartServer(red *redis.Pool, rr redisReceiver, rw redisWriter) {
	client := firebase.NewFirestore()

	app := fiber.New()

	app.Use("/chat", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Static("/", "./static")
	app.Get("/", func(c *fiber.Ctx) error {
		c.WriteString("This works i believe")
		return nil
	})

	app.Get("/chat/:userId", websocket.New(func(c *websocket.Conn) {
		var (
			mt  int
			msg []byte
			err error
		)

		rr.Register(c)

		for {
			mt, msg, err = c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway) || err == io.EOF {
					log.Println("websocket is closed")
					break
				}
				log.Println("Unknown message")
			}

			switch mt {
			case websocket.TextMessage:
				c.WriteMessage(mt, msg)
				if err != nil {
					fmt.Println("Unable to add message to firestore")
				}
				rw.Publish(msg)
			default:
				rw.Publish([]byte("Unknown message"))
			}

		}

		rr.DeRegister(c)

		c.WriteMessage(websocket.CloseMessage, []byte("Websocket closed"))
	}))
	log.Fatal(app.Listen(":9000"))
}
