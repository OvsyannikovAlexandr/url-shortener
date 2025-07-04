package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"url-shortener/internal/storage"
)

func setupTestHandler() *Handler {
	store := storage.NewRedisStorage("localhost:6379")
	if err := store.SaveURL(context.Background(), "abc123", "https://example.com", time.Minute); err != nil {
		log.Fatalf("failed to set up test storage: %v", err)
	}
	return New(store, "", nil)
}

func TestShortenAndRedirect(t *testing.T) {
	h := setupTestHandler()

	// --- Test POST /shorten ---
	payload := []byte(`{"url": "https://test.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	h.Shorten(w, req)

	resp := w.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Ошибка при закрытии тела ответа: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	data, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	shortURL := body["short_url"]
	if shortURL == "" {
		t.Fatalf("short_url not found in response")
	}

	// --- Test GET /{short} ---
	shortKey := shortURL[strings.LastIndex(shortURL, "/")+1:]
	req = httptest.NewRequest(http.MethodGet, "/"+shortKey, nil)
	w = httptest.NewRecorder()
	h.Redirect(w, req)

	resp = w.Result()
	if resp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("expected 301 redirect, got %d", resp.StatusCode)
	}
}
