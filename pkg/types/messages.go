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

type BusyIndicatorMessage struct {
	ChatId int64
}

/*
A chat can be either a group or a user, so we use authuser to attach to each
request message for the purpose of authentication
*/
type AuthUser struct {
	TelegramUserId int64
	ChatName       string
}

type RequestMessage struct {
	UserId   int64
	ChatId   int64
	Message  string
	History  []StoredMessage
	AuthUser AuthUser
	Role     Role
}
