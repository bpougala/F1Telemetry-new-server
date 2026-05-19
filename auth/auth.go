package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

type contextKey string

const KeyIDContextKey contextKey = "keyId"

func NewValkeyClient() *redis.Client {
	endpoint := os.Getenv("VALKEY_ENDPOINT")
	port := os.Getenv("VALKEY_PORT")
	if port == "" {
		port = "6379"
	}
	return redis.NewClient(&redis.Options{
		Addr:      fmt.Sprintf("%s:%s", endpoint, port),
		TLSConfig: &tls.Config{},
	})
}

func Middleware(valkeyClient *redis.Client, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		keyId, err := valkeyClient.Get(r.Context(), "session:"+token).Result()
		if err == redis.Nil {
			http.Error(w, `{"error":"invalid or expired session token"}`, http.StatusUnauthorized)
			return
		}
		if err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), KeyIDContextKey, keyId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ValidateToken(ctx context.Context, valkeyClient *redis.Client, r *http.Request) (string, error) {
	token := ""
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		token = r.URL.Query().Get("token")
	}

	if token == "" {
		return "", fmt.Errorf("no token provided")
	}

	keyId, err := valkeyClient.Get(ctx, "session:"+token).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("invalid or expired session token")
	}
	if err != nil {
		return "", fmt.Errorf("valkey error: %w", err)
	}
	return keyId, nil
}
