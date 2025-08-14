
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
	"strconv"

	"github.com/gorilla/mux"
)

// cache struct
type latestCache struct {
	mu    sync.RWMutex
	data  map[string]map[string]float64 // base -> target -> rate
	since time.Time
}

var latestRates = latestCache{
	data: make(map[string]map[string]float64),
}


// RAW response cache for /historical
type histRespCache struct {
	mu    sync.RWMutex
	data  map[string][]byte        // key -> raw JSON response from exchangerate.host
	since map[string]time.Time     // when we stored it
}

var historicalRespCache = histRespCache{
	data:  make(map[string][]byte),
	since: make(map[string]time.Time),
}


func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// start background refresh for a default base (USD)
	go startHourlyRefresh("USD")

	r := mux.NewRouter()

	// health check
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	r.HandleFunc("/latest", getLatestRate).Methods("GET")
	r.HandleFunc("/convert", convertAmount).Methods("GET")
	r.HandleFunc("/historical", getHistoricalRate).Methods("GET")

	fmt.Println("listening on :" + port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		panic(err)
	}
}

// hourly refresh goroutine
func startHourlyRefresh(base string) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	// run immediately once at startup
	refreshLatestRates(base)

	for range ticker.C {
		refreshLatestRates(base)
	}
}

// fetch & cache latest rates
func refreshLatestRates(base string) {
	apiURL := fmt.Sprintf("https://v6.exchangerate-api.com/v6/b4424af80f21de45e428d97d/latest/%s", base)

	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Println("Error fetching rates:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("API returned error status:", resp.Status)
		return
	}

	var data struct {
		ConversionRates map[string]float64 `json:"conversion_rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("Error decoding API response:", err)
		return
	}

	latestRates.mu.Lock()
	latestRates.data[base] = data.ConversionRates
	latestRates.since = time.Now()
	latestRates.mu.Unlock()

	fmt.Println("Rates updated at", time.Now())
}

// handler: latest
func getLatestRate(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	if base == "" {
		http.Error(w, "missing base query param", http.StatusBadRequest)
		return
	}

	// check cache
	latestRates.mu.RLock()
	rates, ok := latestRates.data[base]
	latestRates.mu.RUnlock()

	if ok {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"base":       base,
			"updated_at": latestRates.since.Format(time.RFC3339),
			"rates":      rates,
		})
		return
	}

	// fetch from API if not in cache
	refreshLatestRates(base)

	latestRates.mu.RLock()
	rates, ok = latestRates.data[base]
	latestRates.mu.RUnlock()

	if !ok {
		http.Error(w, "could not fetch rates", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"base":       base,
		"updated_at": latestRates.since.Format(time.RFC3339),
		"rates":      rates,
	})
}

func convertAmount(w http.ResponseWriter, r *http.Request) {
    from := r.URL.Query().Get("from")
    to := r.URL.Query().Get("to")
    amountStr := r.URL.Query().Get("amount")

    if from == "" || to == "" || amountStr == "" {
        http.Error(w, "missing from, to or amount query param", http.StatusBadRequest)
        return
    }

    amount, err := strconv.ParseFloat(amountStr, 64)
    if err != nil {
        http.Error(w, "invalid amount", http.StatusBadRequest)
        return
    }

    // Lookup rates in the in-memory cache
latestRates.mu.RLock()
fromRates, ok := latestRates.data[from]
latestRates.mu.RUnlock()
  if !ok {
		// Not in cache — fetch from external API
		refreshLatestRates(from)

		// Try cache again
		latestRates.mu.RLock()
		fromRates, ok = latestRates.data[from]
		latestRates.mu.RUnlock()

		if !ok {
			http.Error(w, "could not fetch rates for "+from, http.StatusInternalServerError)
			return
		}
	}

    rate, ok := fromRates[to]
    if !ok {
        http.Error(w, fmt.Sprintf("no rate from %s to %s", from, to), http.StatusNotFound)
        return
    }

    // Do the conversion from cached rate
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












func getHistoricalRate(w http.ResponseWriter, r *http.Request) {
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

	// ---- CACHE CHECK (raw bytes) ----
	cacheKey := fmt.Sprintf("%s|%s|%s", dateStr, source, target)
	historicalRespCache.mu.RLock()
	if raw, ok := historicalRespCache.data[cacheKey]; ok {
		historicalRespCache.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		_, _ = w.Write(raw)
		return
	}
	historicalRespCache.mu.RUnlock()
	// ---- END CACHE CHECK ----

	apiURL := fmt.Sprintf(
		"http://api.exchangerate.host/timeframe?access_key=%s&start_date=%s&end_date=%s&source=%s&currencies=%s&format=1",
		"3fc5b2c59ffad185558841035e114b9c",
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

	// Read once so we can both respond and cache
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read API response", http.StatusInternalServerError)
		return
	}

	// Store in cache
	historicalRespCache.mu.Lock()
	historicalRespCache.data[cacheKey] = raw
	historicalRespCache.since[cacheKey] = time.Now()
	historicalRespCache.mu.Unlock()

	// Return exactly what the API returned (unchanged)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	_, _ = w.Write(raw)
}
