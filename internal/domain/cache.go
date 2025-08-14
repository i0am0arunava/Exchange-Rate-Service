package domain

import (
	"sync"
	"time"
)

type LatestCache struct {
	Mu    sync.RWMutex
	Data  map[string]map[string]float64 // base -> target -> rate
	Since time.Time
}

type HistRespCache struct {
	Mu    sync.RWMutex
	Data  map[string][]byte    // key -> raw JSON response from exchangerate.host
	Since map[string]time.Time // when we stored it
}
