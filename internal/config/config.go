

package config

import (
	"os"
	"github.com/bradfitz/gomemcache/memcache"
	"golang.org/x/sync/singleflight"
)

var (
	MC              *memcache.Client
	SFLatestRates   singleflight.Group
	HistoricalGroup singleflight.Group
)


func InitMemcached() {
	host := os.Getenv("MEMCACHED_HOST")
	if host == "" {
		host = "localhost:11211"
	}
	MC = memcache.New(host)
}
