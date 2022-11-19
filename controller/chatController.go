package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/tsawlen/matchingAppChatService/common/dbInterface"
)

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
