package firebase

import (
	"context"
	"fmt"

	"firebase.google.com/go/auth"
)

type FirebaseAuth struct {
	Client *auth.Client
}

func NewFirebaseAuth() *FirebaseAuth {
	a := NewFirebaseApp()
	auth, err := a.App.Auth(context.Background())
	if err != nil {
		fmt.Errorf("unable to intialize auth client")
	}
	return &FirebaseAuth{
		Client: auth,
	}
}
