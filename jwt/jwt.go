package jwtUtils

import (
	"caddy-auth/cache"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CachedParse(tokenStr string) (jwt.MapClaims, error) {
	c := cache.Storage

	// Check for cache
	if val, found := c.Get(tokenStr); found {
		return val, nil
	}

	// Cache not found, parse it
	claims, err := Parse(tokenStr)
	if err != nil {
		return nil, err
	}

	// Save to cache for n minutes
	c.SetWithTTL(tokenStr, claims, 1, 5*time.Minute)
	return claims, nil
}

func Parse(tokenStr string) (jwt.MapClaims, error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("JWT_KEY")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return nil, err
	}

	// Get claims (data) out of the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("couldn't get claims")
	}

	return claims, nil
}
