package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"url-shortener/internal/service"
	"url-shortener/internal/storage"
)

type Handler struct {
	Storage *storage.RedisStorage
}

func New(storage *storage.RedisStorage) *Handler {
	return &Handler{Storage: storage}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !strings.HasPrefix(req.URL, "http") {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	key := service.GenerateShortKey(6)
	err := h.Storage.SaveURL(r.Context(), key, req.URL, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "ошибка сохранения", http.StatusInternalServerError)
		return
	}

	resp := shortenResponse{ShortURL: fmt.Sprintf("http://localhost:8080/%s", key)}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/")
	if key == "" {
		if _, err := fmt.Fprintln(w, "Добро пожаловать в URL Shortener!"); err != nil {
			log.Println("Ошибка при записи в ответ:", err)
		}
		return
	}

	url, err := h.Storage.GetURL(r.Context(), key)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", h.Shorten)
	mux.HandleFunc("/", h.Redirect)
	return mux
}
