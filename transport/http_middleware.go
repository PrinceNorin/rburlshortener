package transport

import (
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"
)

var (
	rxBearer = regexp.MustCompile(`^Bearer (.+)$`)
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func httpAdminAuthMiddleware(token string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqToken string

			// get token from header
			content := r.Header.Get("Authorization")
			if content != "" {
				if m := rxBearer.FindStringSubmatch(content); len(m) == 2 {
					reqToken = m[1]
				}
			}

			// get token from query string
			content = r.URL.Query().Get("token")
			if content != "" {
				reqToken = content
			}

			if reqToken != token {
				resp := map[string]string{"error": "403 Forbidden!"}
				writeJSON(w, resp, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func loggingMiddleware(logger *log.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := &statusRecorder{
				ResponseWriter: w,
				status:         200,
			}
			defer func(begin time.Time) {
				took := time.Since(begin)
				reqAt := time.Now().Format("2006/01/02 15:04")

				logger.Printf("[%s] %s: %d %v - %s",
					r.Method, reqAt, ww.status, took, r.URL.RequestURI())
			}(time.Now())

			next.ServeHTTP(ww, r)
		})
	}
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// better to log to logging or bug tracking services
				log.Printf("[Error]: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
