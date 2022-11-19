package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/tsawlen/matchingAppChatService/common/database"
	"github.com/tsawlen/matchingAppChatService/controller"
)

func main() {
	sessionChannel := make(chan *gocql.Session)

	go database.InitDB(sessionChannel)

	session := <-sessionChannel

	defer session.Close()

	router := gin.Default()

	// Get Requests
	router.GET("/chats", controller.GetAllChats(session))
	router.GET("/chats/user/:id", controller.GetAllChatsForUser(session))

	router.Run("0.0.0.0:8081")
}
