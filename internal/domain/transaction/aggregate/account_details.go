package aggregate

type AccountByDetails struct {
	Id          uint32  `json:"id"`
	Balance     float32 `json:"balance"`
	AccountName string  `json:"name"`
	Bank        string  `json:"bank"`
	UserId      uint32  `json:"user_id"`
	CreatedAt   string  `json:"created_at"`
}
