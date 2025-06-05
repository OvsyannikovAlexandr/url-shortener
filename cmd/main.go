package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"

	"url-shortener/internal/handler"
	"url-shortener/internal/storage"
)

var templates *template.Template

func main() {
	// Загружаем шаблоны
	templates = template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/index.html",
		"templates/stats.html",
	))

	// Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	store := storage.NewRedisStorage(redisAddr)

	// Хендлеры API
	h := handler.New(store, "https://localhost:8080")

	// Маршруты
	http.HandleFunc("/", showIndex)
	http.HandleFunc("/shorten", h.ShortenHTML)
	http.HandleFunc("/stats", showStats(store))

	// Обработчик редиректа (API)
	http.HandleFunc("/r/", h.Redirect)

	log.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func showIndex(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "layout.html", nil); err != nil {
		http.Error(w, "ошибка шаблона", http.StatusInternalServerError)
	}
}

func showStats(store *storage.RedisStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		data := make(map[string]any)

		if key != "" {
			clicks, err := store.GetClicks(context.Background(), key)
			if err != nil {
				data["Error"] = "Ссылка не найдена или ошибка хранения"
			} else {
				data["Clicks"] = clicks
			}
		}

		if err := templates.ExecuteTemplate(w, "layout.html", data); err != nil {
			http.Error(w, "ошибка шаблона", http.StatusInternalServerError)
		}
	}
}
