package cache

import (
	"github.com/dgraph-io/ristretto/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Global cache storage
var Storage *ristretto.Cache[string, jwt.MapClaims]

func Load() {
	// Example init from github "github.com/dgraph-io/ristretto/v2"
	cache, err := ristretto.NewCache(&ristretto.Config[string, jwt.MapClaims]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})

	if err != nil {
		panic(err)
	}

	Storage = cache
}

func Close() {
	Storage.Close()
}
