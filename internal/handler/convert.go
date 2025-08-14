package handler

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    "exchange-rate-service/internal/domain"
    "exchange-rate-service/internal/service"
)

type ConvertHandler struct {
    Cache *domain.LatestCache
}

func (h *ConvertHandler) ConvertAmount(w http.ResponseWriter, r *http.Request) {
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

    // Lookup rates in cache
    h.Cache.Mu.RLock()
    fromRates, ok := h.Cache.Data[from]
    h.Cache.Mu.RUnlock()

    if !ok {
        // Not in cache — fetch from external API
        service.RefreshLatestRates(from, h.Cache)

        // Try cache again
        h.Cache.Mu.RLock()
        fromRates, ok = h.Cache.Data[from]
        h.Cache.Mu.RUnlock()

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

    // Do conversion
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
