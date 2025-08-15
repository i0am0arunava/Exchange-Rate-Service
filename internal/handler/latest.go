package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"exchange-rate-service/internal/config"
	"exchange-rate-service/internal/service"

)

func GetLatestRate(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	if base == "" {
		http.Error(w, "missing base query param", http.StatusBadRequest)
		return
	}

	cacheKey := fmt.Sprintf("latest:%s", base)

	
	item, err := config.MC.Get(cacheKey)
	if err == nil {
		
		var rates map[string]float64
		if err := json.Unmarshal(item.Value, &rates); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"base":  base,
				"rates": rates,
			})
			return
		}
	}

	
	service.RefreshLatestRates(base)


	
	item, err = config.MC.Get(cacheKey)
	if err != nil {
		http.Error(w, "could not fetch rates", http.StatusInternalServerError)
		return
	}

	var rates map[string]float64
	if err := json.Unmarshal(item.Value, &rates); err != nil {
		http.Error(w, "failed to decode cached rates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"base":  base,
		"rates": rates,
	})
}
