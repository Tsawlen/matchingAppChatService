package dbInterface

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/tsawlen/matchingAppChatService/common/dataStructure"
)

func GetAllMessagesForUser(db *gocql.Session, userId int) (*[]dataStructure.ChatMessage, error) {
	var message dataStructure.ChatMessage
	var messages []dataStructure.ChatMessage

	cnqlQuery1 := "SELECT * FROM chat_space.messages WHERE writtenbyuserid=? ALLOW FILTERING"
	cnqlQuery2 := "SELECT * FROM chat_space.messages WHERE sendtouser=? ALLOW FILTERING"

	iterator1 := db.Query(cnqlQuery1, userId).Iter()
	iterator2 := db.Query(cnqlQuery2, userId).Iter()

	for iterator1.Scan(&message.ChatID, &message.UpdatedAt, &message.ChatID, &message.CreatedAt, &message.Message, &message.Read, &message.SendToUser, &message.WrittenByUserID) {
		messages = append(messages, message)
	}
	if errIterator1 := iterator1.Close(); errIterator1 != nil {
		return nil, errIterator1
	}
	for iterator2.Scan(&message.ChatID, &message.UpdatedAt, &message.ChatID, &message.CreatedAt, &message.Message, &message.Read, &message.SendToUser, &message.WrittenByUserID) {
		messages = append(messages, message)
	}
	if errIterator2 := iterator2.Close(); errIterator2 != nil {
		return nil, errIterator2
	}

	return &messages, nil
}

func SaveMessageToCassandra(db *gocql.Session, message *dataStructure.ChatMessage) error {
	if message.Message == "" {
		return nil
	}
	statement := "INSERT INTO chat_space.messages (messageid, chatid, changedat, createdat, message, read, sendtouser, writtenbyuserid) VALUES (?,?,?,?,?,?,?,?) IF NOT EXISTS"
	messageUUID, errUUID := gocql.RandomUUID()
	if errUUID != nil {
		fmt.Println(errUUID)
	}
	err := db.Query(statement, messageUUID, message.ChatID, message.UpdatedAt, message.CreatedAt, message.Message, message.Read, message.SendToUser, message.WrittenByUserID).Exec()
	if err != nil {
		return err
	}

	return nil
}
