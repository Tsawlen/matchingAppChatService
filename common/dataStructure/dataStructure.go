package dataStructure

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
)

type Chat struct {
	UserId1   int        `json:"userID1"`
	UserId2   int        `json:"userID2"`
	ChatId    gocql.UUID `json:"chatId"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type ChatMessage struct {
	WrittenByUserID int        `json:"writtenBy"`
	SendToUser      int        `json:"sendTo"`
	ChatID          gocql.UUID `json:"chatID"`
	Read            bool       `json:"read"`
	Message         string     `json:"message"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type MessageReceive struct {
	WrittenByUserID int    `json:"writtenBy"`
	SendToUser      int    `json:"sendTo"`
	Read            bool   `json:"read"`
	Message         string `json:"message"`
	Jwt             string `json:"jwt"`
}

type SendMessage struct {
	Message  ChatMessage
	Reciever *websocket.Conn
}

type Login struct {
	UserID int `json:"userID"`
}
