package config

import (
	"github.com/bradfitz/gomemcache/memcache"
	"golang.org/x/sync/singleflight"
)

var (
	MC              = memcache.New("localhost:11211")
	SFLatestRates   singleflight.Group
	HistoricalGroup singleflight.Group
)
