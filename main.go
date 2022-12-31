package main

import (
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/tsawlen/matchingAppChatService/common/database"
	"github.com/tsawlen/matchingAppChatService/common/initializer"
	"github.com/tsawlen/matchingAppChatService/controller"
)

func main() {
	sessionChannel := make(chan *gocql.Session)
	readyChannel := make(chan bool)

	go initializer.LoadEnvVariables(readyChannel)
	<-readyChannel
	go database.InitDB(sessionChannel)
	session := <-sessionChannel

	controller.SetDatabase(session)

	defer session.Close()

	router := mux.NewRouter()

	router.HandleFunc("/getAllMessagesForUser", controller.GetAllChatsForUserMux).Methods("GET")
	router.HandleFunc("/Rooms", controller.GetAllChatRoomsMux).Methods("GET")
	router.HandleFunc("/createChatRoom", controller.CreateChatMux).Methods("PUT")
	router.HandleFunc("/sendMessage", controller.HandleConnections)
	go controller.HandleMessage(session)

	/*if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}*/

	server := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	server.ListenAndServe()

}
