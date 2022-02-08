package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"firebase.google.com/go/auth"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/gofiber/template/html"
	"github.com/gofiber/websocket/v2"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gomodule/redigo/redis"
	"github.com/midepeter/chat-app/firebase"
	"google.golang.org/api/iterator"
)

type MsgData struct {
	Event string
	Data  string
}

func StartServer(red *redis.Pool, rr redisReceiver, rw redisWriter) {
	f := firebase.NewFirestore()

	var msdata MsgData

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
		return c.Render("index", nil)
	})
	app.Post("/auth", handleAuth)

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte("secret"),
	}))

	app.Use(func(c *fiber.Ctx) error {
		token := c.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		uid := claims["uid"].(string)

		a := firebase.NewFirebaseAuth()
		user, err := a.Client.GetUser(context.Background(), uid)
		if err != nil {
			return c.SendStatus(fiber.ErrBadGateway.Code)
		}

		c.Locals("user", user)

		return c.Next()
	})

	app.Get("/rooms", fetchRooms)

	app.Post("/createRoom", createRoom)

	//the ws/:roomID set up a websocket connection to a specific roomID
	app.Get("/ws/:roomID", websocket.New(func(c *websocket.Conn) {

		var (
			mt  int
			msg []byte
			err error
		)
		roomId := c.Params("roomID")

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

			err := json.Unmarshal(msg, &msdata)
			if err != nil {
				errors.New("Unable to unmarshal msgData")
			}

			switch mt {
			case websocket.TextMessage:
				if msdata.Event == "old_messages" {
					var oldMsgData MsgData
					doc, err := f.Client.Collection("chat-app").Doc(roomId).Get(context.Background())
					if err != nil {
						errors.New("unable to fetch previous room messages")
					}

					eleMap := doc.Data()

					var prevmessages []string
					for i, v := range eleMap {
						prevmessages = append(prevmessages, v.(string))
						fmt.Printf("This message sent by %v was %s", i, v)
					}
					for _, j := range prevmessages {
						oldMsgData.Data = j
						c.WriteMessage(mt, []byte(oldMsgData.Data))
					}
				}

				if msdata.Event == "new_messages" {
					_, err = f.Client.Collection("chat-app").Doc(roomId).Set(ctx, msdata.Data)
					if err != nil {
						fmt.Printf("Unable to save message firestore %s", err)
					}

					rw.Publish([]byte(msdata.Data))
				}
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
	Id    string
	Name  string
	Users []*websocket.Conn
}

var rooms []Room

//This function is to get the list of all rooms
func fetchRooms(c *fiber.Ctx) error {
	f := firebase.NewFirestore()

	docItr := f.Client.Collection("chat-app").Documents(context.Background())

	for {
		doc, err := docItr.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return err
		}
		rooms = append(rooms, Room{
			Id: doc.Ref.ID,
		})
	}
	return c.JSON(fiber.Map{
		"rooms": rooms,
	})
}

//CreateRoom route created new room and creates a by creating a doc for in the firstore db
func createRoom(c *fiber.Ctx) error {
	f := firebase.NewFirestore()

	name := c.Body()

	doc := f.Client.Collection("chat-app").NewDoc()

	doc.ID = string(name)

	rooms = append(rooms, Room{
		Id: doc.ID,
	})

	return c.SendString("New room created successfully")
}

//Handles the sign up function of the mobile app
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

//ID token received from front end generated from thee firebase
type IdToken struct {
	Token string `json:"token" form:"token"`
}

//This is to handle autthentication and aslo generates a JWT token which is passed to every
//request that is protected and requires authorization
func handleAuth(c *fiber.Ctx) error {
	a := firebase.NewFirebaseAuth()
	var idToken IdToken
	if err := c.BodyParser(&idToken); err != nil {
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
		return c.SendStatus(fiber.ErrBadGateway.Code)
	}

	claims := jwt.MapClaims{
		"uid":   uid,
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := jwtToken.SignedString([]byte("secret"))
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	return c.JSON(fiber.Map{"token": t})
}
