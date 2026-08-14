package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const RequestIDHeader = "X-Request-ID"

type requestIDContextKey struct{}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = newRequestID()
		}

		ctx := context.WithValue(
			r.Context(),
			requestIDContextKey{},
			requestID,
		)

		w.Header().Set(RequestIDHeader, requestID)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func newRequestID() string {
	var buffer [16]byte

	if _, err := rand.Read(buffer[:]); err == nil {
		return hex.EncodeToString(buffer[:])
	}

	return fmt.Sprintf(
		"req-%d",
		time.Now().UnixNano(),
	)
}
