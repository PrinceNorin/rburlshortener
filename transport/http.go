package transport

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewHTTPHandler() http.Handler {
	r := mux.NewRouter()

	return r
}
