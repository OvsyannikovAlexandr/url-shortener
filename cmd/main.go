package main

import (
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
	templates = template.Must(template.ParseGlob("templates/*"))

	// Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379" //"redis:6379" "localhost:6379"
	}
	store := storage.NewRedisStorage(redisAddr)

	// Хендлеры API
	h := handler.New(store, "https://localhost:8080", templates)

	// Маршруты
	http.HandleFunc("/shorten", h.ShortenHTML)
	http.HandleFunc("/stats", h.ShowStats)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" || path == "/index.html" {
			h.ShowIndex(w, r)
			return
		}
		h.Redirect(w, r)
	})

	log.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// func showIndex(w http.ResponseWriter, r *http.Request) {
// 	if err := template.Must(template.ParseFiles(
// 		"templates/layout.html",
// 		"templates/index.html",
// 	)).ExecuteTemplate(w, "layout.html", nil); err != nil {
// 		http.Error(w, "ошибка шаблона", http.StatusInternalServerError)
// 	}
// }
