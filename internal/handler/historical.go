package handler

import (

	"fmt"
	"io"
	"net/http"
	"time"
    "os"
	"log"
	"exchange-rate-service/internal/config"

	"github.com/bradfitz/gomemcache/memcache"
)



func GetHistoricalRate(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	source := r.URL.Query().Get("source")
	target := r.URL.Query().Get("target")

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

	// Memcached cache key
	cacheKey := fmt.Sprintf("historical:%s|%s|%s", dateStr, source, target)

	// Try Memcached first
	if item, err := config.MC.Get(cacheKey); err == nil {
		// Cache hit
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		_, _ = w.Write(item.Value)
		return
	} else if err != memcache.ErrCacheMiss {
		http.Error(w, "error fetching from cache: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Use singleflight to avoid duplicate API fetch for same key
	v, err, _ := config.HistoricalGroup.Do(cacheKey, func() (interface{}, error) {
		// Check cache again inside singleflight (another goroutine might have set it)
		if item, err := config.MC.Get(cacheKey); err == nil {
			return item.Value, nil
		}
        
		// Fetch from API

		 historicalBaseURL := os.Getenv("HISTORICAL_API_BASE_URL")
         historicalAPIKey := os.Getenv("HISTORICAL_API_KEY")

         if historicalBaseURL == "" || historicalAPIKey == "" {
         log.Fatal("HISTORICAL_API_BASE_URL or HISTORICAL_API_KEY is not set in environment variables")
         }
		apiURL := fmt.Sprintf(
       "%s?access_key=%s&start_date=%s&end_date=%s&source=%s&currencies=%s&format=1",
        historicalBaseURL,
        historicalAPIKey,
       dateStr, dateStr, source, target,
       )

		resp, err := http.Get(apiURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("external API error: %s", resp.Status)
		}

		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// Store in Memcached for 24 hours
		if err := config.MC.Set(&memcache.Item{
			Key:        cacheKey,
			Value:      raw,
			Expiration: 86400,
		}); err != nil {
			fmt.Println("Warning: could not set Memcached:", err)
		}

		return raw, nil
	})

	if err != nil {
		http.Error(w, "failed to fetch historical rate: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	_, _ = w.Write(v.([]byte))
}
