package dbInterface

import (
	"github.com/gocql/gocql"
	"github.com/tsawlen/matchingAppChatService/common/dataStructure"
)

func GetAllChats(session *gocql.Session) (*[]dataStructure.Chat, error) {
	var chat dataStructure.Chat
	var chats []dataStructure.Chat

	cnqlQuery := "SELECT * FROM chat_space.chat"
	iterator := session.Query(cnqlQuery).Iter()
	for iterator.Scan(&chat.UserId1, &chat.UserId2, &chat.CreatedAt, &chat.UpdatedAt) {
		chats = append(chats, chat)
	}

	if errIterator := iterator.Close(); errIterator != nil {
		return nil, errIterator
	}
	return &chats, nil
}

func GetAllChatsForUser(session *gocql.Session, userId int) (*[]dataStructure.Chat, error) {
	var chat dataStructure.Chat
	var chats []dataStructure.Chat

	cnqlQuery1 := "SELECT * FROM chat_space.chat WHERE userid1=?"
	cnqlQuery2 := "SELECT * FROM chat_space.chat WHERE userid2=? ALLOW FILTERING"
	iterator1 := session.Query(cnqlQuery1, userId).Iter()
	iterator2 := session.Query(cnqlQuery2, userId).Iter()
	for iterator1.Scan(&chat.UserId1, &chat.UserId2, &chat.Active, &chat.UpdatedAt, &chat.ChatId, &chat.CreatedAt) {
		chats = append(chats, chat)
	}
	if errIterator1 := iterator1.Close(); errIterator1 != nil {
		return nil, errIterator1
	}
	for iterator2.Scan(&chat.UserId1, &chat.UserId2, &chat.Active, &chat.UpdatedAt, &chat.ChatId, &chat.CreatedAt) {
		chats = append(chats, chat)
	}
	if errIterator2 := iterator2.Close(); errIterator2 != nil {
		return nil, errIterator2
	}
	return &chats, nil
}
