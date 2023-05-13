package layers

import (
	"log"
	"ratatoskr/types"
	"testing"
)

type mockCapability struct {
	checkScore   float32
	response     types.ResponseMessage
	executeError error
}

func newMockCapability(checkScore float32, response types.ResponseMessage, executeError error) *mockCapability {
	return &mockCapability{
		checkScore:   checkScore,
		response:     response,
		executeError: executeError,
	}
}

func (m *mockCapability) Check(req *types.RequestMessage) float32 {
	return m.checkScore
}

func (m *mockCapability) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {
	return m.response, m.executeError
}

func makeMockCapability(checkScore float32, executeResponse types.ResponseMessage, executeError error) types.Capability {
	return newMockCapability(checkScore, executeResponse, executeError)
}

func Test(t *testing.T) {

	// Arrange
	caps := []types.Capability{
		makeMockCapability(1.0, types.ResponseMessage{Message: "correct choice"}, nil),
		makeMockCapability(1.0, types.ResponseMessage{Message: "incorrect choice"}, nil),
		makeMockCapability(0.4, types.ResponseMessage{Message: "incorrect choice"}, nil),
	}

	capsel := NewCapabilitySelector(caps)
	log.Printf("%v", capsel)

	rm := types.RequestMessage{}

	// Act
	res, err := capsel.PassThrough(&rm)

	// Assert
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if res.Message != "correct choice" {
		t.Errorf("Expected 'correct choice', got '%v'", res.Message)
	}

}
