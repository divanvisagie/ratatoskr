package layers

import (
	"fmt"
	"testing"
	"time"
)

func TestLimitedMemoryLenght(t *testing.T) {
	memoryLayer := NewMemoryLayer(nil)
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 2) //we need to wait because unix timestamp is a key
		memoryLayer.saveRequestMessage("test_user", fmt.Sprintf("test message %d", i))
	}

	messages := memoryLayer.getMessages("test_user")

	if len(messages) != 20 {
		t.Errorf("Expected 10, got %d", len(memoryLayer.store["test"]))
	}
}
