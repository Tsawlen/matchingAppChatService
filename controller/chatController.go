package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
	"github.com/tsawlen/matchingAppChatService/common/dataStructure"
	"github.com/tsawlen/matchingAppChatService/common/dbInterface"
	"github.com/tsawlen/matchingAppChatService/middleware"
)

var connectedClients = make(map[int]*websocket.Conn)
var broadcaster = make(chan dataStructure.MessageReceive)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(request *http.Request) bool {
		return true
	},
}
var db *gocql.Session

// REST section

func GetAllChats(session *gocql.Session) gin.HandlerFunc {
	handler := func(context *gin.Context) {
		chats, err := dbInterface.GetAllChats(session)
		if err != nil {
			fmt.Println(err)
			context.AbortWithStatusJSON(http.StatusNoContent, gin.H{
				"error": err,
			})
			return
		}
		context.JSON(http.StatusOK, chats)
	}
	return gin.HandlerFunc(handler)
}

func GetAllChatsForUser(session *gocql.Session) gin.HandlerFunc {
	handler := func(context *gin.Context) {

		userId, errConv := strconv.Atoi(context.Param("id"))
		if errConv != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "User ID must be a number",
			})
			return
		}
		chats, err := dbInterface.GetAllChatsForUser(session, userId)
		if err != nil {
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}
		context.JSON(http.StatusOK, chats)
	}
	return gin.HandlerFunc(handler)
}

func GetAllChatsForUserMux(w http.ResponseWriter, r *http.Request) {
	var userId = r.Header.Get("user")
	var authorization = r.Header.Get("authorization")
	intUser, errConv := strconv.Atoi(userId)
	if errConv != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
	authorized, errAuth := middleware.Auth(authorization, int(intUser))
	if errAuth != nil {
		http.Error(w, errAuth.Error(), http.StatusInternalServerError)
	}
	if !authorized {
		http.Error(w, "Unauthorized!", http.StatusUnauthorized)
	}
	messages, err := dbInterface.GetAllMessagesForUser(db, intUser)
	if err != nil {
		http.Error(w, errAuth.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(messages)

}

// Websocket Section

func HandleConnections(writer http.ResponseWriter, request *http.Request) {
	newWebSocket, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer newWebSocket.Close()

	for {
		var msg dataStructure.MessageReceive
		errReadLoop := newWebSocket.ReadJSON(&msg)
		_, ok := connectedClients[msg.WrittenByUserID]
		if !ok {
			connectedClients[msg.WrittenByUserID] = newWebSocket
		}
		authorized, errAuth := middleware.Auth(msg.Jwt, msg.WrittenByUserID)
		if errAuth != nil {
			log.Println("Error validating user: " + errAuth.Error())
		}
		if !authorized {
			user, err := getCorrectConnectionToClose(newWebSocket)
			if err != nil {
				log.Println("User socket could not be deleted!")
			}
			delete(connectedClients, user)
			break
		}
		if errReadLoop != nil {
			user, err := getCorrectConnectionToClose(newWebSocket)
			if err != nil {
				log.Println("User socket could not be deleted!")
			}
			delete(connectedClients, user)
			break
		}
		broadcaster <- msg
	}
}

func HandleMessage(db *gocql.Session) {
	for {
		msg := <-broadcaster

		_, err := prepareSendMessage(db, &msg)
		if err != nil {
			fmt.Println("Error sending message: " + err.Error())
		}
	}
}

func saveChatMessage(db *gocql.Session, chatMessage *dataStructure.ChatMessage) error {
	if err := dbInterface.SaveMessageToCassandra(db, chatMessage); err != nil {
		return err
	}
	return nil
}

func getChatRoomToSocket(db *gocql.Session, userId int) (*websocket.Conn, bool) {
	socket, ok := connectedClients[userId]
	if ok {
		return socket, true
	}
	return nil, false

}

func chatRoomExists(db *gocql.Session, sender int, receiver int) (gocql.UUID, bool, error) {
	allChatsForSender, errGetChats := dbInterface.GetAllChatsForUser(db, sender)
	newUUID, errCreateUUID := gocql.RandomUUID()
	if errCreateUUID != nil {
		fmt.Println("Failed to create a uuid!")
		return newUUID, false, nil
	}
	if errGetChats != nil {
		return newUUID, false, errGetChats
	}

	for _, data := range *allChatsForSender {
		if data.UserId2 == receiver || data.UserId1 == receiver {
			return data.ChatId, true, nil
		}
	}

	return newUUID, false, nil
}

func sendMessage(client *websocket.Conn, message *dataStructure.ChatMessage) {
	err := client.WriteJSON(message)
	if err != nil && websocket.IsCloseError(err, websocket.CloseGoingAway) {
		fmt.Println("Error: " + err.Error())
		client.Close()
		delete(connectedClients, message.SendToUser)
	}
}

func prepareSendMessage(db *gocql.Session, message *dataStructure.MessageReceive) (bool, error) {
	socket, online := getChatRoomToSocket(db, message.SendToUser)
	chatUUID, chatExists, errSearchChat := chatRoomExists(db, message.WrittenByUserID, message.SendToUser)
	messageToSave := convertToMessage(message)
	if errSearchChat != nil {
		return false, errSearchChat
	}
	if !chatExists {
		if err := dbInterface.CreateNewChatForUsers(db, message.WrittenByUserID, message.SendToUser); err != nil {
			log.Println(err)
		}
	}
	messageToSave.ChatID = chatUUID
	errSave := saveChatMessage(db, messageToSave)
	if errSave != nil {
		return false, errSave
	}
	if online {
		sendMessage(socket, messageToSave)
	}
	return true, nil
}

// Helper functions

func correctMessage(message *dataStructure.ChatMessage, chatUUID gocql.UUID) *dataStructure.ChatMessage {
	message.ChatID = chatUUID
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()
	return message
}

func convertToMessage(msg *dataStructure.MessageReceive) *dataStructure.ChatMessage {
	currentTime := time.Now()
	message := dataStructure.ChatMessage{
		WrittenByUserID: msg.WrittenByUserID,
		SendToUser:      msg.SendToUser,
		Read:            false,
		Message:         msg.Message,
		CreatedAt:       currentTime,
		UpdatedAt:       currentTime,
	}
	return &message
}

func getCorrectConnectionToClose(socket *websocket.Conn) (int, error) {
	for counter, data := range connectedClients {
		if reflect.DeepEqual(data, socket) {
			return counter, nil
		}
	}
	return 0, errors.New("No connected used found for this socket!")
}

func SetDatabase(session *gocql.Session) {
	db = session
}
