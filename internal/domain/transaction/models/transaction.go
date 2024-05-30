package models

import (
	"time"
)

var (
	TRANSACTIONTYPEEXPECTS = []string{TRANSACTIONTYPEDEPOSIT, TRANSACTIONTYPEWITHDRAW}
)

var (
	TRANSACTIONTYPEDEPOSIT  = "deposit"
	TRANSACTIONTYPEWITHDRAW = "withdraw"
)

var TRANSACTIONTABLE = "transactions"
var (
	TRANSACTIONCOLUMN_ID               = TRANSACTIONTABLE + ".id"
	TRANSACTIONCOLUMN_ACCOUNT_ID       = TRANSACTIONTABLE + ".account_id"
	TRANSACTIONCOLUMN_AMOUNT           = TRANSACTIONTABLE + ".amount"
	TRANSACTIONCOLUMN_TRANSACTION_TYPE = TRANSACTIONTABLE + ".transaction_type"
	TRANSACTIONCOLUMN_CREATED_AT       = TRANSACTIONTABLE + ".created_at"
	TRANSACTIONCOLUMN_UPDATED_AT       = TRANSACTIONTABLE + ".updated_at"
	TRANSACTIONCOLUMN_DELETED          = TRANSACTIONTABLE + ".deleted"
)

type Transaction struct {
	ID              uint32    `gorm:"column:id;primaryKey;autoIncrement;not null"`
	AccountID       uint32    `gorm:"column:account_id;not null"`
	Amount          float32   `gorm:"column:amount;not null"`
	TransactionType string    `gorm:"column:transaction_type;type:varchar(15);not null"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
	Deleted         bool      `gorm:"column:deleted;index"`
}
