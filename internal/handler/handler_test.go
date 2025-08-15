
package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"testing"

	"exchange-rate-service/internal/config"
	"exchange-rate-service/internal/handler"

	"github.com/bradfitz/gomemcache/memcache"
)


func TestGetLatestRate_CacheHit(t *testing.T) {
	
	config.MC = memcache.New("localhost:11211")

	
	cacheKey := "latest:USD"
	rates := map[string]float64{"EUR": 0.85, "INR": 83.0}
	data, _ := json.Marshal(rates)
	err := config.MC.Set(&memcache.Item{Key: cacheKey, Value: data, Expiration: 30})
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	req := httptest.NewRequest("GET", "/latest?base=KID", nil)
	w := httptest.NewRecorder()

	handler.GetLatestRate(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}


}

func TestGetLatestRate_MissingBase(t *testing.T) {
	config.MC = memcache.New("localhost:11211")

	req := httptest.NewRequest("GET", "/latest", nil) // no ?base param
	w := httptest.NewRecorder()

	handler.GetLatestRate(w, req)

	if w.Code != 400 {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestConvertAmount_LatestCacheHit(t *testing.T) {
	
	cacheKey := "latest:USD"
	rates := map[string]float64{"INR": 83.0}
	data, _ := json.Marshal(rates)
	_ = config.MC.Set(&memcache.Item{Key: cacheKey, Value: data})

	req := httptest.NewRequest("GET", "/convert?from=USD&to=INR&amount=2", nil)
	w := httptest.NewRecorder()

	handler.ConvertAmount(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	
}



func TestGetHistoricalRate_CacheHit(t *testing.T) {
	config.MC = memcache.New("localhost:11211")

	cacheKey := "historical:2025-08-01|USD|INR"
	data, _ := json.Marshal(map[string]interface{}{"rate": 83.0})
	if err := config.MC.Set(&memcache.Item{Key: cacheKey, Value: data, Expiration: 30}); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	req := httptest.NewRequest("GET", "/historical?date=2025-08-01&source=USD&target=INR", nil)
	w := httptest.NewRecorder()

	handler.GetHistoricalRate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("X-Cache") != "HIT" {
		t.Errorf("expected cache hit, got %s", w.Header().Get("X-Cache"))
	}
}
