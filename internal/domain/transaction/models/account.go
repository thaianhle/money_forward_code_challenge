package models

import (
	"gorm.io/gorm"
	"time"
)

var (
	BANKACBTYPE = "ACB"
	BANKVCBTYPE = "VCB"
	BANKVIBTYPE = "VIB"
)
var (
	BANKEXPECTEDS = []string{BANKACBTYPE, BANKVCBTYPE, BANKVIBTYPE}
)

var ACCOUNTTABLE = "accounts"
var (
	ACCOUNTCOLUMN_ID         = ACCOUNTTABLE + "." + "id"
	ACCOUNTCOLUMN_BANK       = ACCOUNTTABLE + "." + "bank"
	ACCOUNTCOLUMN_BALANCE    = ACCOUNTTABLE + "." + "balance"
	ACCOUNTCOLUMN_NAME       = ACCOUNTTABLE + "." + "name"
	ACCOUNTCOLUMN_USER_ID    = ACCOUNTTABLE + "." + "user_id"
	ACCOUNTCOLUMN_CREATED_AT = ACCOUNTTABLE + "." + "created_at"
	ACCOUNTCOLUMN_UPDATED_AT = ACCOUNTTABLE + "." + "updated_at"
	ACCOUNTCOLUMN_DELETED_AT = ACCOUNTTABLE + "." + "deleted"
)

type Account struct {
	ID        uint32         `gorm:"column:id;primaryKey;autoIncrement;not null"`
	Bank      string         `gorm:"column:bank;type:char(3);not null"`
	Balance   float32        `gorm:"column:balance;not null"`
	Name      string         `gorm:"column:name;type:varchar(255);not null"`
	UserId    uint32         `gorm:"column:user_id;not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	Deleted   gorm.DeletedAt `gorm:"colum:deleted;index"`
}
