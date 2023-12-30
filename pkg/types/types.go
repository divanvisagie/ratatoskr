package types

type ResponseMessage struct {
	ChatID  int64
	Message string
	Bytes   []byte
	Context []StoredMessage
}

type RequestMessage struct {
	UserName string
	ChatID   int64
	Message  string
	Context  []StoredMessage
}

type StoredMessage struct {
	Role      string    `json:"role"`
	Message   string    `json:"message"`
	Timestamp int64     `json:"timestamp"`
	Embedding []float32 `json:"embedding"`
	Rank      float32   `json:"rank"`
}

type MessageOnDisk struct {
	Role    string `yaml:"role"`
	Content string `yaml:"content"`
	Hash    string `yaml:"hash"`
}

type Capability interface {
	Check(req *RequestMessage) float32
	Execute(req *RequestMessage) (ResponseMessage, error)
}
