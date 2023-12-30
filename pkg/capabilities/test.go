package caps

import "ratatoskr/pkg/types"


type Test struct {
}

func NewTest() *Test {
	return &Test{}
}

func (t Test) Check(req *types.RequestMessage) float32 {
	if req.Message == "test" {
		return 1
	}
	return 0
}

func (t Test) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {
	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: "This is a test response to reduce the amount of gpt calls",
	}
	return res, nil
}
