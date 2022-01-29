package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"firebase.google.com/go/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/gofiber/websocket/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/midepeter/chat-app/firebase"
	uuid "github.com/nu7hatch/gouuid"
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
		c.Render("chat-app", fiber.Map{
			"User": "olumide",
		})
		return nil
	})
	app.Post("/login", handleLogin)
	app.Get("/retrieve", handleRetrieve)
	app.Post("/createRoom", createRoom)
	app.Post("/joinRoom", joinRoom)
	app.Get("/fetch", fetchAllRooms)

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {

		var (
			mt  int
			msg []byte
			err error
		)
		roomId := c.Params("room_id")

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

				_, err := f.Client.Collection("chat-app").Doc(roomId).Set(ctx, storemsg)
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

type Room struct {
	Id   string
	Name string
	User []*websocket.Conn
}

var rooms []Room

func NewRoom(name string) *Room {
	u, err := uuid.NewV4()
	if err != nil {
		fmt.Println("return err")
	}
	return &Room{
		Id:   u.String(),
		Name: name,
	}
}

//This function is to get the list of all rooms
func fetchAllRooms(c *fiber.Ctx) error {
	f := firebase.NewFirestore()
	docItr := f.Client.Collection("chat-app").Documents(context.Background())

	docs, err := docItr.GetAll()
	if err != nil {
		fmt.Errorf("There was an error fething all documents from firestore %s", err)
		return err
	}
	user := c.Body()
	fmt.Println(docs)
	log.Println(user)
	c.Render("chatroooms", fiber.Map{
		//"User": string(user),
		//"Docs": docs,
	})
	return nil
}

//createRoom initializes a new room where other users can join in to chat
func createRoom(c *fiber.Ctx) error {
	f := firebase.NewFirestore()

	name := c.Body()

	//new := NewRoom(string(name))

	//rooms = append(rooms, &new)
	doc := f.Client.Collection("chat-app").NewDoc()
	doc.ID = string(name)
	fmt.Printf("A new room was successfully created %s", doc.ID)
	return nil
}

func joinRoom(c *fiber.Ctx) error {
	a := firebase.NewFirebaseAuth()
	var idToken IdToken
	if err := c.BodyParser(&idToken); err != nil {
		fmt.Println(err)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	token, err := a.Client.VerifyIDToken(context.Background(), idToken.Token)
	if err != nil {
		fmt.Printf("Was unable to verify id token %v", err)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	uid := token.UID
	user, err := a.Client.GetUser(context.Background(), uid)
	if err != nil {
		fmt.Println(err)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	c.Render("chat-app", fiber.Map{
		"User": user,
	})
	return nil
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

	_, err := a.Client.CreateUser(context.Background(), params)
	if err != nil {
		fmt.Errorf("Unable to create user %v", err)
	}
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

	c.Redirect("/fetch")
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
	fmt.Println(eleMap)
	return nil
}
