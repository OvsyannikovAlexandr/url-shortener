package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"url-shortener/internal/handler"
	"url-shortener/internal/storage"
)

func TestShortenAndRedirect(t *testing.T) {
	// Настраиваем Redis
	redisURL := os.Getenv("REDIS_ADDR")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	store := storage.NewRedisStorage(redisURL)
	// Поднимаем тестовый сервер
	h := handler.New(store, "")
	server := httptest.NewServer(h.Router())
	defer server.Close()

	h.BaseURL = server.URL

	// 1. Тестируем POST /shorten
	originalURL := "https://example.com"
	body := map[string]string{"url": originalURL}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(server.URL+"/shorten", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		t.Fatalf("POST /shorten failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Ошибка при закрытии тела ответа: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var respData map[string]string
	data, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(data, &respData); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	shortURL := respData["short_url"]

	// 2. Тестируем GET /{short_code}
	time.Sleep(100 * time.Millisecond) // Redis может быть чуть медленным в тесте

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // не следовать за редиректом
		},
	}

	getResp, err := client.Get(shortURL)
	if err != nil {
		t.Fatalf("GET short URL failed: %v", err)
	}
	// defer getResp.Body.Close()
	defer func() {
		if err := getResp.Body.Close(); err != nil {
			t.Logf("Ошибка при закрытии: %v", err)
		}
	}()

	if getResp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("expected 301 redirect, got %d", getResp.StatusCode)
	}

	location := getResp.Header.Get("Location")
	if location != originalURL {
		t.Fatalf("expected redirect to %s, got %s", originalURL, location)
	}
}
