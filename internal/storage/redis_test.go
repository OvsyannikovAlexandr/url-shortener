package storage

import (
	"context"
	"os"
	"testing"
	"time"
)

var testAddr = "localhost:6379"

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestRedisStorage_SaveAndGet(t *testing.T) {
	storage := NewRedisStorage(testAddr)
	ctx := context.Background()

	key := "testKey"
	val := "https://example.com"
	ttl := time.Minute

	err := storage.SaveURL(ctx, key, val, ttl)
	if err != nil {
		t.Fatalf("failed to save url: %v", err)
	}

	got, err := storage.GetURL(ctx, key)
	if err != nil {
		t.Fatalf("failed to get url: %v", err)
	}

	if got != val {
		t.Errorf("expected %s, got %s", val, got)
	}
}
