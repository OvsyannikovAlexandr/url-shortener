package main

import (
	"context"
	"log"
	"net/http"

	"url-shortener/internal/handler"
	"url-shortener/internal/storage"
)

func main() {
	ctx := context.Background()
	storage := storage.NewRedisStorage("redis:6379")

	if err := storage.Ping(ctx); err != nil {
		log.Fatalf("Redis не отвечает: %v", err)
	}

	h := handler.New(storage, "")

	http.HandleFunc("/shorten", h.Shorten)
	http.HandleFunc("/", h.Redirect)
	http.HandleFunc("/stats/", h.Stats)

	log.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
