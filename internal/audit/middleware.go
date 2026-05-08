package audit

import (
	"fmt"
	"net/http"
)

// Middleware wraps an http.Handler and records each incoming request
// as a ConfigLoad audit event (reused as an API-access event).
func Middleware(log *Log, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		log.Record(EventConfigLoad, "", msg)
		next.ServeHTTP(w, r)
	})
}
