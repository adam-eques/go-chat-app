package internal

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

type Message struct {
	User      string `json:"user"`
	Data      []byte `json:"data"`
	TimeStamp time.Time
}

func validate(data []byte) (Message, error) {
	var msg Message

	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, errors.Wrap(err, "Cannot handle message")
	}

	if msg.User == "" {
		return msg, errors.New("Message has not user nor data")
	}

	return msg, nil
}
