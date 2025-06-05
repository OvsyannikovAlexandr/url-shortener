package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"url-shortener/internal/service"
	"url-shortener/internal/storage"
)

type Handler struct {
	Storage *storage.RedisStorage
	BaseURL string
}

func New(storage *storage.RedisStorage, baseUrl string) *Handler {
	return &Handler{
		Storage: storage,
		BaseURL: baseUrl,
	}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

type statsResponse struct {
	URL    string `json:"url"`
	Clicks int64  `json:"clicks"`
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

	resp := shortenResponse{ShortURL: fmt.Sprintf("%s/%s", h.BaseURL, key)}
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

	if err := h.Storage.IncrementClicks(r.Context(), key); err != nil {
		log.Println("Ошибка при увеличении счётчика переходов:", err)
	}

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", h.Shorten)
	mux.HandleFunc("/", h.Redirect)
	mux.HandleFunc("/stats/", h.Stats)
	return mux
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/stats/")
	if key == "" {
		http.Error(w, "ключ не указан", http.StatusBadRequest)
		return
	}

	url, err := h.Storage.GetURL(r.Context(), key)
	if err != nil {
		http.Error(w, "ссылка не найдена", http.StatusNotFound)
		return
	}

	clicks, err := h.Storage.GetClicks(r.Context(), key)
	if err != nil {
		clicks = 0 // если нет значения — считаем 0
	}

	resp := statsResponse{
		URL:    url,
		Clicks: clicks,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "ошибка ответа", http.StatusInternalServerError)
	}
}

func (h *Handler) ShortenHTML(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	key := service.GenerateShortKey(6)
	err := h.Storage.SaveURL(r.Context(), key, url, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "ошибка сохранения", http.StatusInternalServerError)
		return
	}

	shortURL := "http://localhost:8080/r/" + key
	data := map[string]any{"ShortURL": shortURL}
	if err := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/index.html",
	)).ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, "ошибка шаблона", http.StatusInternalServerError)
	}

}
