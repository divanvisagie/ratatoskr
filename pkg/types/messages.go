package types

type StoredMessage struct {
	ChatId int64
	Content string
}

type ResponseMessage struct {
	ChatId int64
	Message string
}

type RequestMessage struct {
	ChatId int64
	Message string 
}
