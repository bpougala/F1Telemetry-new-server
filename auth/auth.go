package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler/transport"
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
		// Skip auth for WebSocket upgrade requests (iOS URLSessionWebSocketTask
		// cannot set custom headers on the handshake request)
		if isWebSocketUpgrade(r) {
			next.ServeHTTP(w, r)
			return
		}

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

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

// WebsocketInitFunc validates the auth token sent in the connection_init payload.
// iOS URLSessionWebSocketTask cannot set HTTP headers on the handshake, so the
// client sends {"Authorization": "Bearer <token>"} in the connection_init payload.
func WebsocketInitFunc(valkeyClient *redis.Client) transport.WebsocketInitFunc {
	return func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
		token := initPayload.Authorization()
		if token == "" {
			return ctx, nil, fmt.Errorf("missing or invalid Authorization in connection_init payload")
		}

		keyId, err := valkeyClient.Get(ctx, "session:"+token).Result()
		if err == redis.Nil {
			return ctx, nil, fmt.Errorf("invalid or expired session token")
		}
		if err != nil {
			return ctx, nil, fmt.Errorf("internal server error")
		}

		ctx = context.WithValue(ctx, KeyIDContextKey, keyId)
		return ctx, nil, nil
	}
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
