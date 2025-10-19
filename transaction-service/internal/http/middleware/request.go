package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

const headerRequestID = "X-Request-ID"

// RequestID ensures each request has an X-Request-ID
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rid := r.Header.Get(headerRequestID); rid != "" {
			w.Header().Set(headerRequestID, rid)
			next.ServeHTTP(w, r)
			return
		}
		id := uuid.NewString()
		w.Header().Set(headerRequestID, id)
		r.Header.Set(headerRequestID, id)
		next.ServeHTTP(w, r)
	})
}
