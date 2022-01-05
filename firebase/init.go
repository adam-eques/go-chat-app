package firebase

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
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
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}
	return &Firebaseapp{
		App: app,
	}
}
