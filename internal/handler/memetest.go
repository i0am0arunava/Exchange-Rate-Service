package handler

import (
	"encoding/json"
	"net/http"

	"exchange-rate-service/internal/config"

	"github.com/bradfitz/gomemcache/memcache"
)

// MemTestHandler tests storing and retrieving a value from Memcached
func MemTestHandler(w http.ResponseWriter, r *http.Request) {
	key := "test-key"
	value := []byte("Hello from Memcached!")

	// Store in Memcached for 30 seconds
	if err := config.MC.Set(&memcache.Item{Key: key, Value: value, Expiration: 30}); err != nil {
		http.Error(w, "Set failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve from Memcached
	item, err := config.MC.Get(key)
	if err != nil {
		http.Error(w, "Get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"stored":  string(value),
		"fetched": string(item.Value),
	})
}
