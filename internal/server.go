package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"firebase.google.com/go/auth"
	"github.com/acentior/chat-app/firebase"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/gofiber/websocket/v2"
	"github.com/gomodule/redigo/redis"
)

func StartServer(red *redis.Pool, rr redisReceiver, rw redisWriter) {
	f := firebase.NewFirestore()

	http_port := os.Getenv("PORT")
	fmt.Println(http_port)

	ctx := context.Background()
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Static("/", "./static")
	app.Get("/Signup", handleSignup)
	app.Get("/", func(c *fiber.Ctx) error {
		c.Render("index", nil)
		return nil
	})
	app.Get("/chat", func(c *fiber.Ctx) error {
		c.Render("chat-app", nil)
		return nil
	})
	app.Post("/login", handleLogin)
	// app.Get("/retrieve", handleRetrieve)

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
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
				storemsg := Message{
					User: c.Params("id"),
					Data: msg,
				}
				message := "Echo: " + string(msg)
				c.WriteMessage(mt, []byte(message))

				_, err := f.Client.Collection("chat-app").Doc("messages").Set(ctx, storemsg)
				if err != nil {
					fmt.Println("Unable to save message firestore")
				}

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
	log.Fatal(app.Listen(fmt.Sprintf(":%s", http_port)))
}

type User struct {
	Username string
	Password string
}

func handleSignup(c *fiber.Ctx) error {
	a := firebase.NewFirebaseAuth()
	var user User
	if err := c.BodyParser(&user); err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	params := (&auth.UserToCreate{}).
		Email(user.Username).
		EmailVerified(false).
		Password(user.Password)

	u, err := a.Client.CreateUser(context.Background(), params)
	if err != nil {
		fmt.Errorf("Unable to create user %v", err)
	}

	fmt.Println("successfully create the user", u)
	return nil
}

type IdToken struct {
	Token string `json:"token" form:"token"`
}

func handleLogin(c *fiber.Ctx) error {
	a := firebase.NewFirebaseAuth()
	var idToken IdToken
	if err := c.BodyParser(&idToken); err != nil {
		fmt.Println(err)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	token, err := a.Client.VerifyIDToken(context.Background(), idToken.Token)
	if err != nil {
		fmt.Println(err)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	uid := token.UID
	user, err := a.Client.GetUser(context.Background(), uid)
	if err != nil {
		fmt.Println(err)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	fmt.Println(user)

	c.Render("chat-app", fiber.Map{
		"User": user,
	})
	return nil
}

func handleRetrieve(c *fiber.Ctx) error {
	f := firebase.NewFirestore()

	doc, err := f.Client.Collection("chat-app").Doc("messages").Get(context.Background())
	if err != nil {
		return errors.New("Could not retrieve document")
	}

	eleMap := doc.Data()
	for i, v := range eleMap {
		fmt.Printf("This message sent by %v was %s", i, v)
	}
	return nil
}
