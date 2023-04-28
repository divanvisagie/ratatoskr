package types

type ResponseMessage struct {
	ChatID  int64
	Message string
}

type RequestMessage struct {
	ChatID  int64
	Message string
}

type StoredMessage struct {
	Role    string
	Message string
}
