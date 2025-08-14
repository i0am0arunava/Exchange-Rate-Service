package main

import (
	"fmt"
	"net/http"
	"os"
	"log"
    "time"
	"github.com/bradfitz/gomemcache/memcache"
	"exchange-rate-service/internal/config" 
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"exchange-rate-service/internal/service"
    "exchange-rate-service/internal/handler"
)





func init() {
	// Load .env file into environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using system env variables")
	}

	// Optional: check if important env vars are loaded
	if os.Getenv("API_BASE_URL") == "" {
		log.Fatal("API_BASE_URL is not set in env")
	}
	if os.Getenv("HISTORICAL_API_BASE_URL") == "" {
		log.Fatal("HISTORICAL_API_BASE_URL is not set in env")
	}
	if os.Getenv("HISTORICAL_API_KEY") == "" {
		log.Fatal("HISTORICAL_API_KEY is not set in env")
	}

	fmt.Println("Environment variables loaded successfully")
}







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
