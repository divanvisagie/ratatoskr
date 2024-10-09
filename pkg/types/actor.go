package types

type Actor interface {
	SendMessage(RequestMessage)
	ReceiveMessage() ResponseMessage 
	Stop()
}
