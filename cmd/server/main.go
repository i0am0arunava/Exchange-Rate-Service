
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
	// "encoding/json"
    "io"
	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := mux.NewRouter()

	// health check
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// placeholder endpoints
	r.HandleFunc("/latest", getLatestRate).Methods("GET")
	r.HandleFunc("/convert", convertAmount).Methods("GET")
	r.HandleFunc("/historical", getHistoricalRate).Methods("GET")

	fmt.Println("listening on :" + port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		panic(err)
	}
}

// handlers — for now, just placeholders
func getLatestRate(w http.ResponseWriter, r *http.Request) {
    base := r.URL.Query().Get("base")
    if base == "" {
        http.Error(w, "missing base query param", http.StatusBadRequest)
        return
    }

    apiURL := fmt.Sprintf("https://v6.exchangerate-api.com/v6/b4424af80f21de45e428d97d/latest/%s", base)

    resp, err := http.Get(apiURL)
    if err != nil {
        http.Error(w, "failed to fetch rates", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        http.Error(w, "external API error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if _, err := io.Copy(w, resp.Body); err != nil {
        http.Error(w, "failed to read API response", http.StatusInternalServerError)
    }
}


func convertAmount(w http.ResponseWriter, r *http.Request) {
    from := r.URL.Query().Get("from")
    to := r.URL.Query().Get("to")
    amount := r.URL.Query().Get("amount")

    if from == "" || to == "" || amount == "" {
        http.Error(w, "missing from, to or amount query param", http.StatusBadRequest)
        return
    }

    apiURL := fmt.Sprintf(
        "https://v6.exchangerate-api.com/v6/b4424af80f21de45e428d97d/pair/%s/%s/%s",
        from, to, amount,
    )

    resp, err := http.Get(apiURL)
    if err != nil {
        http.Error(w, "failed to fetch conversion", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        http.Error(w, "external API error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if _, err := io.Copy(w, resp.Body); err != nil {
        http.Error(w, "failed to read API response", http.StatusInternalServerError)
    }
}



func getHistoricalRate(w http.ResponseWriter, r *http.Request) {
    dateStr := r.URL.Query().Get("date")
    source := r.URL.Query().Get("source")
    target := r.URL.Query().Get("target")

    if dateStr == "" || source == "" || target == "" {
        http.Error(w, "missing date, source, or target query param", http.StatusBadRequest)
        return
    }

    // Parse the date
    reqDate, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        http.Error(w, "invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
        return
    }

    // Check if older than 90 days
    if time.Since(reqDate).Hours() > 90*24 {
        http.Error(w, "date exceeds 90-day history limit", http.StatusBadRequest)
        return
    }

    // Build API URL
    apiURL := fmt.Sprintf(
        "http://api.exchangerate.host/timeframe?access_key=%s&start_date=%s&end_date=%s&source=%s&currencies=%s&format=1",
        "3fc5b2c59ffad185558841035e114b9c",
        dateStr, dateStr, source, target,
    )

    resp, err := http.Get(apiURL)
    if err != nil {
        http.Error(w, "failed to fetch historical rate", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        http.Error(w, "external API error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if _, err := io.Copy(w, resp.Body); err != nil {
        http.Error(w, "failed to read API response", http.StatusInternalServerError)
    }
}
