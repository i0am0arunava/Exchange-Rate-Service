package handler

import (
	"fmt"
	"io"
	"net/http"
	"time"
     "os"
	"exchange-rate-service/internal/domain"
)

type HistoricalHandler struct {
	Cache *domain.HistRespCache
}

func (h *HistoricalHandler) GetHistoricalRate(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	source := r.URL.Query().Get("source")
	target := r.URL.Query().Get("target")
 

	apiBaseURL := os.Getenv("HISTORICAL_API_BASE_URL")
    apiKey := os.Getenv("HISTORICAL_API_KEY")
    if apiBaseURL == "" || apiKey == "" {
	http.Error(w, "historical API configuration missing", http.StatusInternalServerError)
	return
    }


	if dateStr == "" || source == "" || target == "" {
		http.Error(w, "missing date, source, or target query param", http.StatusBadRequest)
		return
	}

	reqDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	if time.Since(reqDate).Hours() > 90*24 {
		http.Error(w, "date exceeds 90-day history limit", http.StatusBadRequest)
		return
	}

	// ---- CACHE CHECK (raw bytes) ----
	cacheKey := fmt.Sprintf("%s|%s|%s", dateStr, source, target)
	h.Cache.Mu.RLock()
	if raw, ok := h.Cache.Data[cacheKey]; ok {
		h.Cache.Mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		_, _ = w.Write(raw)
		return
	}
	h.Cache.Mu.RUnlock()
	// ---- END CACHE CHECK ----

	apiURL := fmt.Sprintf(
	"%s?access_key=%s&start_date=%s&end_date=%s&source=%s&currencies=%s&format=1",
	apiBaseURL,
	apiKey,
	dateStr, dateStr, source, target,
)

	resp, err := http.Get(apiURL)
	if err != nil {
		http.Error(w, "failed to fetch historical rate", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "external API error", http.StatusInternalServerError)
		return
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read API response", http.StatusInternalServerError)
		return
	}

	// Store in cache
	h.Cache.Mu.Lock()
	h.Cache.Data[cacheKey] = raw
	h.Cache.Since[cacheKey] = time.Now()
	h.Cache.Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	_, _ = w.Write(raw)
}
