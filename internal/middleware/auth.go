package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
)

func AuthMiddleware(secretKey string, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")

			var userID int64

			if err == nil && cookie.Value != "" {
				parsedUserID, valid := VerifyToken(cookie.Value, secretKey)
				if valid {
					userID = parsedUserID
				}
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GenerateToken(userID int64, secretKey string) (string, error) {
	timestamp := time.Now().Unix()
	payload := fmt.Sprintf("%d.%d", userID, timestamp)

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(payload))
	signature := hex.EncodeToString(h.Sum(nil))

	return payload + "." + signature, nil
}

func VerifyToken(token, secretKey string) (int64, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, false
	}

	payload := parts[0] + "." + parts[1]
	receivedSignature := parts[2]

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(receivedSignature), []byte(expectedSignature)) {
		return 0, false
	}

	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, false
	}

	return userID, true
}

func GetUserIDFromContext(ctx context.Context) int64 {
	value := ctx.Value(userIDKey)
	if value == nil {
		return 0
	}

	userID, ok := value.(int64)
	if !ok {
		return 0
	}
	return userID
}
