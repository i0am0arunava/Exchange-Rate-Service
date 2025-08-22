package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
    "exchange-rate-service/internal/config"
	"github.com/bradfitz/gomemcache/memcache"
	
)


func RefreshLatestRates(base string) {
	   

   
	 baseURL := "https://v6.exchangerate-api.com/v6/dd8a2660664f756430ea7b73"
     

	   // Use singleflight to ensure only one API fetch for the same base currency
      // is performed at a time, avoiding duplicate concurrent requests.
	  
	_, err, _ := config.SFLatestRates.Do(base, func() (interface{}, error) {
		apiURL := fmt.Sprintf("%s/latest/%s", baseURL, base)

		resp, err := http.Get(apiURL)
		if err != nil {
			fmt.Println("Error fetching rates:", err)
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("API returned error status:", resp.Status)
			return nil, fmt.Errorf("API status: %s", resp.Status)
		}

		var data struct {
			ConversionRates map[string]float64 `json:"conversion_rates"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			fmt.Println("Error decoding API response:", err)
			return nil, err
		}

		ratesJSON, err := json.Marshal(data.ConversionRates)
		if err != nil {
			fmt.Println("Error marshaling rates:", err)
			return nil, err
		}

		err = config.MC.Set(&memcache.Item{
			Key:        fmt.Sprintf("latest:%s", base),
			Value:      ratesJSON,
			Expiration: 3600,
		})
		if err != nil {
			fmt.Println("Error setting Memcached:", err)
			return nil, err
		}

		fmt.Println("Rates updated in Memcached for base:", base, "at", time.Now())
		return nil, nil
	})

	if err != nil {
		fmt.Println("singleflight refresh failed:", err)
	}
}
