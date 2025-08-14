package handler

import (
    "encoding/json"
    "net/http"
    "time"

    "exchange-rate-service/internal/domain"
    "exchange-rate-service/internal/service"
)

type LatestHandler struct {
    Cache *domain.LatestCache
}

func (h *LatestHandler) GetLatestRate(w http.ResponseWriter, r *http.Request) {
    base := r.URL.Query().Get("base")
    if base == "" {
        http.Error(w, "missing base query param", http.StatusBadRequest)
        return
    }

    h.Cache.Mu.RLock()
    rates, ok := h.Cache.Data[base]
    h.Cache.Mu.RUnlock()

    if ok {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "base":       base,
            "updated_at": h.Cache.Since.Format(time.RFC3339),
            "rates":      rates,
        })
        return
    }

    service.RefreshLatestRates(base, h.Cache)

    h.Cache.Mu.RLock()
    rates, ok = h.Cache.Data[base]
    h.Cache.Mu.RUnlock()

    if !ok {
        http.Error(w, "could not fetch rates", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "base":       base,
        "updated_at": h.Cache.Since.Format(time.RFC3339),
        "rates":      rates,
    })
}
