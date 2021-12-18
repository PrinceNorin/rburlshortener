package transport

import (
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

var (
	rxBearer = regexp.MustCompile(`^Bearer (.+)$`)
)

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
