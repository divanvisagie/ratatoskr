package types

type ResponseMessage struct {
	ChatID  int64
	Message string
}

type RequestMessage struct {
	UserName string
	ChatID   int64
	Message  string
	Context  []StoredMessage
}

type StoredMessage struct {
	Role    string
	Message string
}

type Capability interface {
	Check(req *RequestMessage) bool
	Execute(req *RequestMessage) (ResponseMessage, error)
}
