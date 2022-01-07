package firebase

import (
	"context"
	"fmt"

	firestore "cloud.google.com/go/firestore"
)

type Firestore struct {
	Client *firestore.Client
}

func NewFirestore() Firestore {
	a := NewFirebaseApp()
	ctx := context.Background()
	client, err := a.App.Firestore(ctx)
	if err != nil {
		fmt.Println("Unable to initialize firestore")
	}
	return Firestore{Client: client}
}

/* func (f *Firestore) Add(ctx context.Context, c string, msg []byte) error {
	_, err := f.client.Collection(c).Doc("messages").Set(ctx, map[string]interface{}{
		"message": msg,
	})

	if err != nil {
		fmt.Println("Unable to add message to store")
	}

	return nil
}
*/
