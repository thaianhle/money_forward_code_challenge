package models

import (
	"time"
)

var (
	USERTABLE            = "users"
	USERCOLUMN_ID        = USERTABLE + ".id"
	USERCOLUMN_FIRSTNAME = USERTABLE + ".first_name"
	USERCOLUMN_LASTNAME  = USERTABLE + ".last_name"
	USERCOLUMN_CREATEDAT = USERTABLE + ".created_at"
	USERCOLUMN_UPDATEDAT = USERTABLE + ".updated_at"
	USERCOLUMN_DELETEDAT = USERTABLE + ".deleted"
)

type User struct {
	ID        uint32    `gorm:"column:id;primaryKey;autoIncrement;not null"`
	FirstName string    `gorm:"column:first_name;type:varchar(50);not null"`
	LastName  string    `gorm:"column:last_name;type:varchar(50);not full"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
	Deleted   bool      `gorm:"column:deleted;index"`
}
