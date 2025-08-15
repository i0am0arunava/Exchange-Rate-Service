package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
    "os"
    "log"
	"exchange-rate-service/internal/config"
	"exchange-rate-service/internal/service"

	"github.com/bradfitz/gomemcache/memcache"
)


func ConvertAmount(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	amountStr := r.URL.Query().Get("amount")
	dateStr := r.URL.Query().Get("date")

        if dateStr != "" {
    cacheKey := fmt.Sprintf("historical:%s:%s:%s:%s", from, to, amountStr, dateStr)

   
    if item, err := config.MC.Get(cacheKey); err == nil {
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("X-Cache", "HIT")
        w.Write(item.Value)
        return
    }

    historicalBaseURL := os.Getenv("HISTORICAL_API_BASE_URL")
    historicalAPIKey := os.Getenv("HISTORICAL_API_KEY")

    if historicalBaseURL == "" || historicalAPIKey == "" {
        log.Fatal("HISTORICAL_API_BASE_URL or HISTORICAL_API_KEY is not set in environment variables")
    }

   
    apiURL := fmt.Sprintf(
        "https://api.exchangerate.host/convert?access_key=2ac108b461f57948c41e61ff6a0e210f&from=%s&to=%s&amount=%s&format=1&date=%s",
        from, to, amountStr, dateStr,
    )

    resp, err := http.Get(apiURL)
    if err != nil {
        http.Error(w, "failed to fetch historical conversion: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        http.Error(w, "failed to read API response: "+err.Error(), http.StatusInternalServerError)
        return
    }

    
    var parsed map[string]interface{}
    if err := json.Unmarshal(body, &parsed); err != nil {
        http.Error(w, "invalid API response format", http.StatusInternalServerError)
        return
    }

    if success, ok := parsed["success"].(bool); ok && success {
        
        if err := config.MC.Set(&memcache.Item{Key: cacheKey, Value: body, Expiration: 86400}); err != nil {
            log.Println("cache set failed:", err)
        }
    } else {
       
        log.Println("API returned success=false, skipping cache")
    }

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Cache", "MISS")
    w.Write(body)
    return
}

	
	if from == "" || to == "" || amountStr == "" {
		http.Error(w, "missing from, to or amount query param", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}

	cacheKey := fmt.Sprintf("latest:%s", from)
	item, err := config.MC.Get(cacheKey)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			
			service.RefreshLatestRates(from)

			item, err = config.MC.Get(cacheKey)
			if err != nil {
				http.Error(w, "could not fetch rates for "+from, http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "error fetching from cache: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var fromRates map[string]float64
	if err := json.Unmarshal(item.Value, &fromRates); err != nil {
		http.Error(w, "error decoding cached rates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rate, ok := fromRates[to]
	if !ok {
		http.Error(w, fmt.Sprintf("no rate from %s to %s", from, to), http.StatusNotFound)
		return
	}

	result := amount * rate

	json.NewEncoder(w).Encode(map[string]interface{}{
		"from":      from,
		"to":        to,
		"amount":    amount,
		"rate":      rate,
		"result":    result,
		"cache_hit": true,
	})
}
