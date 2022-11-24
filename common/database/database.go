package database

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/tsawlen/matchingAppChatService/common/mockData"
)

func InitDB(sessionChannel chan *gocql.Session) (*gocql.Session, error) {
	cluster := gocql.NewCluster("localhost:9042")
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = (time.Second * 40)
	session, err := cluster.CreateSession()

	if err != nil {
		fmt.Println("Connection to Cluster failed!")
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("Connected to Cassandra!")

	if errKeyspaceCreate := createKeySpace(session); errKeyspaceCreate != nil {
		fmt.Println("Error creating Keyspace!")
		return nil, errKeyspaceCreate
	}
	if errTableCreate := createChatTableIfNotExists(session); errTableCreate != nil {
		fmt.Println("Error creating Table chats!")
		return nil, errTableCreate
	}
	if errTableCreate := createMessagesTableIfNotExists(session); errTableCreate != nil {
		fmt.Println("Error creating Table messages!")
		return nil, errTableCreate
	}

	if errMockData := insertMockData(session); errMockData != nil {
		fmt.Println("Error inserting Mockdata!")
		return nil, errMockData
	}

	sessionChannel <- session

	return session, nil
}

func createKeySpace(session *gocql.Session) error {
	err := session.Query("CREATE KEYSPACE IF NOT EXISTS chat_space WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 3}").Exec()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func createChatTableIfNotExists(session *gocql.Session) error {
	err := session.Query("CREATE TABLE IF NOT EXISTS chat_space.chat(userid1 int, userid2 int, chatId uuid, active boolean, createdAt timestamp, changedAt timestamp, PRIMARY KEY(userid1, userid2) )").Exec()
	if err != nil {
		return err
	}
	return nil
}

func createMessagesTableIfNotExists(session *gocql.Session) error {
	err := session.Query("CREATE TABLE IF NOT EXISTS chat_space.messages(messageId uuid, writtenbyuserid int, sendtouser int, chatid uuid, read boolean, message text, createdAt timestamp, changedAt timestamp, PRIMARY KEY(messageId))").Exec()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func insertMockData(session *gocql.Session) error {
	uuid, _ := gocql.RandomUUID()
	if err := session.Query("INSERT INTO chat_space.chat (userid1, userid2, chatId, active, createdAt, changedAt) VALUES (?,?,?,?,?,?) IF NOT EXISTS",
		mockData.MockChatData[0].UserId1,
		mockData.MockChatData[0].UserId2,
		uuid,
		true,
		mockData.MockChatData[0].CreatedAt,
		mockData.MockChatData[0].UpdatedAt).Exec(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
