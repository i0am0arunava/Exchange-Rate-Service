package main

import (
	"fmt"
	"net/http"
	"os"
    "time"
	"github.com/bradfitz/gomemcache/memcache"
	"exchange-rate-service/internal/config" 
	"github.com/gorilla/mux"
	"exchange-rate-service/internal/service"
    "exchange-rate-service/internal/handler")

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
    config.MC = memcache.New("localhost:11211")
    r := mux.NewRouter()
    r.HandleFunc("/latest", handler.GetLatestRate).Methods("GET")
    r.HandleFunc("/convert", handler.ConvertAmount).Methods("GET")
    r.HandleFunc("/historical", handler.GetHistoricalRate).Methods("GET")
    r.HandleFunc("/mem-test", handler.MemTestHandler).Methods("GET")

	
    go startHourlyRefresh("USD") 
	fmt.Println("Listening on :" + port)
	http.ListenAndServe(":"+port, r)
}


// hourly refresh goroutine
func startHourlyRefresh(base string) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	 service.RefreshLatestRates(base)
	 for range ticker.C {
		service.RefreshLatestRates(base)

	}
}
