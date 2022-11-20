package main

import (
	"log"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/tsawlen/matchingAppChatService/common/database"
	"github.com/tsawlen/matchingAppChatService/controller"
)

func main() {
	sessionChannel := make(chan *gocql.Session)

	go database.InitDB(sessionChannel)

	session := <-sessionChannel

	defer session.Close()

	http.HandleFunc("/sendMessage", controller.HandleConnections)
	go controller.HandleMessage(session)

	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}

}
