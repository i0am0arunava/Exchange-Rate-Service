
package main

import (
	
	"fmt"

	"net/http"
	"os"
	
	"time"

    "exchange-rate-service/internal/domain"
	"exchange-rate-service/internal/service"
    "exchange-rate-service/internal/handler"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)



var latestRates = domain.LatestCache{
    Data: make(map[string]map[string]float64),
}

var historicalRespCache = domain.HistRespCache{
    Data:  make(map[string][]byte),
    Since: make(map[string]time.Time),
}



func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// start background refresh for a default base (USD)
	go startHourlyRefresh("USD")

	r := mux.NewRouter()
    _ = godotenv.Load() 
	// health check
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	latestHandler := &handler.LatestHandler{Cache: &latestRates}
    r.HandleFunc("/latest", latestHandler.GetLatestRate).Methods("GET")



	convertHandler := &handler.ConvertHandler{Cache: &latestRates}
    r.HandleFunc("/convert", convertHandler.ConvertAmount).Methods("GET")


	historicalHandler := &handler.HistoricalHandler{Cache: &historicalRespCache}
    r.HandleFunc("/historical", historicalHandler.GetHistoricalRate).Methods("GET")


	fmt.Println("listening on :" + port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		panic(err)
	}
}

// hourly refresh goroutine
func startHourlyRefresh(base string) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	
	service.RefreshLatestRates(base, &latestRates)

	for range ticker.C {
		service.RefreshLatestRates(base, &latestRates)
	}
}












