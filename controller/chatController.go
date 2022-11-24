package controller

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
	"github.com/tsawlen/matchingAppChatService/common/dataStructure"
	"github.com/tsawlen/matchingAppChatService/common/dbInterface"
)

var connectedClients = make(map[int]*websocket.Conn)
var broadcaster = make(chan dataStructure.ChatMessage)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(request *http.Request) bool {
		return true
	},
}

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

func HandleConnections(writer http.ResponseWriter, request *http.Request) {
	newWebSocket, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer newWebSocket.Close()

	for {
		var msg dataStructure.ChatMessage
		errReadLoop := newWebSocket.ReadJSON(&msg)
		_, ok := connectedClients[msg.WrittenByUserID]
		if !ok {
			connectedClients[msg.WrittenByUserID] = newWebSocket
		} else {
			if errReadLoop != nil {
				delete(connectedClients, msg.WrittenByUserID)
				break
			}

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

func prepareSendMessage(db *gocql.Session, message *dataStructure.ChatMessage) (bool, error) {
	socket, online := getChatRoomToSocket(db, message.SendToUser)
	chatUUID, chatExists, errSearchChat := chatRoomExists(db, message.WrittenByUserID, message.SendToUser)
	if errSearchChat != nil {
		return false, errSearchChat
	}
	if !chatExists {
		// Create new Chat in DB
	}
	message.ChatID = chatUUID
	errSave := saveChatMessage(db, message)
	if errSave != nil {
		return false, errSave
	}
	if online {
		sendMessage(socket, message)
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
