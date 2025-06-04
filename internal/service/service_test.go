package service

import (
	"testing"
)

func TestGenerateShortKey(t *testing.T) {
	key := GenerateShortKey(6)
	
	if len(key) != 6 {
		t.Errorf("Expected key length of 6, got %d", len(key))
	}

	for _, char := range key {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') {
			t.Errorf("Generated key contains invalid character: %c", char)
		}
	}
}