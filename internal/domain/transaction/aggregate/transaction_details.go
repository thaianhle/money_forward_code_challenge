package aggregate

import (
	"fmt"
	"time"
)

type TransactionByDetails struct {
	// field of transaction model
	Id              uint32  `json:"id"`
	Amount          float32 `json:"amount"`
	TransactionType string  `json:"transaction_type"`
	CreatedAt       string  `json:"created_at"`
	// field of account model
	AccountId uint32 `json:"account_id"`
	Bank      string `json:"bank"`

	// field of user id
	UserId uint32 `json:"user_id"`
}

var (
	ASIA_VN_TIMEZONE = "Asia/Bangkok"
)

// default layout date
var DefaultLayoutDate string = "2006-01-02 15:04:05 +0700 UTC"
var DefaultLayoutDateUTCMST string = "2006-01-02 15:04:05.000 -0700 MST"

func (t *TransactionByDetails) FormatDateHCM(layout ...string) error {
	l := DefaultLayoutDate

	if len(layout) > 0 {
		l = layout[0]
	}

	fmt.Println("time: ", t.CreatedAt)

	timeObj, err := time.Parse(DefaultLayoutDateUTCMST, t.CreatedAt)
	if err != nil {
		timeObj, err = time.Parse(time.RFC3339, t.CreatedAt)
		if err != nil {
			return err
		}
	}

	// Load the location for +0700 timezone
	loc := time.FixedZone("UTC+7", 7*60*60) // +07:00 timezone offset
	convertedTime := timeObj.In(loc)

	t.CreatedAt = convertedTime.Format(l)
	return nil
}
