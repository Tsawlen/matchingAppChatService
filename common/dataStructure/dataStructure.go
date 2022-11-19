package dataStructure

import "time"

type Chat struct {
	UserId1   int       `json:"userID1"`
	UserId2   int       `json:"userID2"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
