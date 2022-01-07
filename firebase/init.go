package firebase

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var (
	projectid string = "chat-app-e2953"
)

type Firebaseapp struct {
	App *firebase.App
}

func NewFirebaseApp() *Firebaseapp {
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: projectid}
	opt := option.WithCredentialsFile("chat-app-e2953-firebase-adminsdk-vjtvt-8710fa5a0a.json")
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		log.Fatalln(err)
	}
	return &Firebaseapp{
		App: app,
	}
}
