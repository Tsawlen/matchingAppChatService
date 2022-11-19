package mockData

import (
	"fmt"
	"time"

	"github.com/tsawlen/matchingAppChatService/common/dataStructure"
)

var MockChatData = []dataStructure.Chat{
	{
		UserId1: 1, UserId2: 2, Active: true, CreatedAt: stringToTime("2022-11-19 14:02:00.000"), UpdatedAt: stringToTime("2022-11-19 14:02:00.000"),
	},
}

func stringToTime(dateString string) time.Time {
	dateStringBlueprint := "2022-11-19 14:02:00.000"
	date, err := time.Parse(dateStringBlueprint, dateString)
	if err != nil {
		fmt.Println("Convertion failed!")
		return time.Now()
	}
	return date
}
