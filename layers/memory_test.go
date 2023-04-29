package layers

import (
	"fmt"
	"testing"
	"time"
)

func TestOnlyReturnsLast20WhenMoreThan20(t *testing.T) {
	memoryLayer := NewMemoryLayer(nil)
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 2) //we need to wait because unix timestamp is a key
		memoryLayer.saveRequestMessage("test_user", fmt.Sprintf("test message %d", i))
	}

	messages := memoryLayer.getMessages("test_user")

	if len(messages) != 20 {
		t.Errorf("Expected 20, got %d", len(memoryLayer.store["test"]))
	}

	//test that memory management happened
	if len(memoryLayer.store["test_user"]) != 10 {
		t.Errorf("Expected 100, got %d", len(memoryLayer.store["test"]))
	}
}

func TestOnlyReturnsLast20WhenLessThan20(t *testing.T) {
	memoryLayer := NewMemoryLayer(nil)
	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond * 2) //we need to wait because unix timestamp is a key
		memoryLayer.saveRequestMessage("test_user", fmt.Sprintf("test message %d", i))
	}

	messages := memoryLayer.getMessages("test_user")

	if len(messages) != 10 {
		t.Errorf("Expected 10, got %d", len(memoryLayer.store["test"]))
	}
}
