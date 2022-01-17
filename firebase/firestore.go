package firebase

import (
	"context"
	"fmt"

	firestore "cloud.google.com/go/firestore"
)

type Firestore struct {
	Client *firestore.Client
}

//initialiing the firestore
func NewFirestore() Firestore {
	a := NewFirebaseApp()
	ctx := context.Background()
	client, err := a.App.Firestore(ctx)
	if err != nil {
		fmt.Println("Unable to initialize firestore: ", err)
	}
	return Firestore{Client: client}
}
