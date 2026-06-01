package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	healthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "healthy" {
		t.Errorf("expected status=healthy, got %s", resp["status"])
	}
}

func TestShortenHandler_Success(t *testing.T) {
	body := `{"url": "https://github.com/Amarachi-Ezeonyekwere"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	shortenHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}

	var resp ShortenResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.ShortCode == "" {
		t.Error("expected a short_code in response")
	}
	if resp.OriginalURL != "https://github.com/Amarachi-Ezeonyekwere" {
		t.Errorf("unexpected original_url: %s", resp.OriginalURL)
	}
}

func TestShortenHandler_InvalidURL(t *testing.T) {
	body := `{"url": "not-a-valid-url"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	shortenHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestShortenHandler_MissingURL(t *testing.T) {
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	shortenHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestShortenHandler_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/shorten", nil)
	rr := httptest.NewRecorder()

	shortenHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestRedirectHandler_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rr := httptest.NewRecorder()

	redirectHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestRedirectHandler_Success(t *testing.T) {
	// First shorten a URL to get a code
	store.mu.Lock()
	store.urls["testcd"] = "https://example.com"
	store.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/testcd", nil)
	rr := httptest.NewRecorder()

	redirectHandler(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("expected status 302, got %d", rr.Code)
	}
	if rr.Header().Get("Location") != "https://example.com" {
		t.Errorf("unexpected redirect location: %s", rr.Header().Get("Location"))
	}
}