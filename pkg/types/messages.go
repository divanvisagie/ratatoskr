package types

type StoredMessage struct {
	CreatedAt int64  `json:"created_at"`
	Content   string `json:"content"`
	Role      string `json:"role"`
}

type ResponseMessage struct {
	UserId  int64
	ChatId  int64
	Message string
}

type RequestMessage struct {
	UserId  int64
	ChatId  int64
	Message string
	History []StoredMessage
}
