package firebase

import (
	"context"
	"fmt"

	firestore "cloud.google.com/go/firestore"
)

type Firestore struct {
	client *firestore.Client
}

func NewFirestore() Firestore {
	a := NewFirebaseApp()
	client, err := a.App.Firestore(context.Background())
	if err != nil {
		fmt.Println("Unable to initialize firestore")
	}
	return Firestore{client: client}
}

func (f *Firestore) Add(ctx context.Context, c string, msg []byte) error {
	_, err := f.client.Collection(c).Doc("messages").Set(ctx, map[string]interface{}{
		"message": msg,
	})

	if err != nil {
		fmt.Println("Unable to add message to store")
	}

	return nil
}
