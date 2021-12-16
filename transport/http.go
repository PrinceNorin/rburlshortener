package transport

import (
	"net/http"

	"github.com/PrinceNorin/rburlshortener/service"
	"github.com/gorilla/mux"
)

// NewHTTPHandler factory function
func NewHTTPHandler(svc service.URLShortener) http.Handler {
	r := mux.NewRouter()

	return r
}
