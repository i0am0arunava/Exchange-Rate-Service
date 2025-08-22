package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
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
	fromDate := r.URL.Query().Get("fromDate")
	toDate := r.URL.Query().Get("toDate")
     


     


 if(fromDate!="" && toDate!=""){

	if toDate =="" || fromDate == "" || from == "" || to == "" {
		http.Error(w, "missing date, source, or target query param", http.StatusBadRequest)
		return
	}

	reqDate, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		http.Error(w, "invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	if time.Since(reqDate).Hours() > 90*24 {
		http.Error(w, "date exceeds 90-day history limit", http.StatusBadRequest)
		return
	}

	
	cacheKey := fmt.Sprintf("historical:%s|%s|%s|%s", fromDate,toDate, from, to)

	
	if item, err := config.MC.Get(cacheKey); err == nil {
		
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
        
		

		 historicalBaseURL := os.Getenv("HISTORICAL_API_BASE_URL")
         historicalAPIKey := os.Getenv("HISTORICAL_API_KEY")

         if historicalBaseURL == "" || historicalAPIKey == "" {
         log.Fatal("HISTORICAL_API_BASE_URL or HISTORICAL_API_KEY is not set in environment variables")
         }
		apiURL := fmt.Sprintf(
       "%s?access_key=%s&start_date=%s&end_date=%s&source=%s&currencies=%s&format=1",
        historicalBaseURL,
        historicalAPIKey,
       fromDate, toDate, from, to,
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

	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	_, _ = w.Write(v.([]byte))

		  }



        




    if fromDate != "" {
       cacheKey := fmt.Sprintf("historical:%s:%s:%s:%s", from, to, amountStr, fromDate)

   
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
        "https://api.exchangerate.host/convert?access_key=84c55c279cdb130c67a3c3992c21f24c&from=%s&to=%s&amount=%s&format=1&date=%s",
        from, to, amountStr, fromDate,
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
