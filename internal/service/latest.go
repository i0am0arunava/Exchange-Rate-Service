package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
     "os"
	"exchange-rate-service/internal/domain"
)



func RefreshLatestRates(base string,cache *domain.LatestCache) {

	apiBaseURL := os.Getenv("API_BASE_URL")
	if apiBaseURL == "" {
		fmt.Println("API_BASE_URL not set")
		return
	}
	apiURL := fmt.Sprintf("%s/latest/%s", apiBaseURL, base)

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

	cache.Mu.Lock()
	cache.Data[base] = data.ConversionRates
	cache.Since = time.Now()
	cache.Mu.Unlock()

	fmt.Println("Rates updated at", time.Now())
}