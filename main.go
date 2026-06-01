package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// URLStore holds the in-memory map of short codes to original URLs
type URLStore struct {
	mu   sync.RWMutex
	urls map[string]string
}

// ShortenRequest is the expected JSON body for POST /shorten
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse is the JSON response for a successful shorten
type ShortenResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// ErrorResponse is the standard error JSON response
type ErrorResponse struct {
	Error string `json:"error"`
}

var store = &URLStore{
	urls: make(map[string]string),
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

// healthHandler returns a simple 200 OK for health checks
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// shortenHandler accepts a URL and returns a short code
func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid or missing 'url' field"})
		return
	}

	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "url must start with http:// or https://"})
		return
	}

	code := generateCode(6)

	store.mu.Lock()
	store.urls[code] = req.URL
	store.mu.Unlock()

	host := os.Getenv("APP_HOST")
	if host == "" {
		host = "http://localhost:8080"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ShortenResponse{
		ShortCode:   code,
		ShortURL:    fmt.Sprintf("%s/%s", host, code),
		OriginalURL: req.URL,
	})
}

// redirectHandler looks up the short code and redirects
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")

	store.mu.RLock()
	originalURL, exists := store.urls[code]
	store.mu.RUnlock()

	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "short code not found"})
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/shorten", shortenHandler)
	mux.HandleFunc("/", redirectHandler)

	log.Printf("🚀 URL Shortener running on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}