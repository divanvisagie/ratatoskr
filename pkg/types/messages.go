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

	// If the response is a file, it will be sent to the user as a document and the text will be used as the filename
	Data []byte
}

type RequestMessage struct {
	UserId  int64
	ChatId  int64
	Message string
	History []StoredMessage
}
