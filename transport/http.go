package transport

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PrinceNorin/rburlshortener/service"
	"github.com/gorilla/mux"
)

var (
	jsonContentType = "application/json"
)

// Config server configuration
type HTTPConfig struct {
	Service    service.URLShortener
	ServerHost string
}

// NewHTTPHandler factory function
func NewHTTPHandler(conf HTTPConfig) http.Handler {
	r := mux.NewRouter()
	h := handler{svc: conf.Service, serverHost: conf.ServerHost}

	r.HandleFunc("/", h.createShortURL).
		Methods("POST")

	return r
}

// internal type definition & implementation
type httpError interface {
	StatusCode() int
	Response() interface{}
}

type createRequest struct {
	URL       string `json:"url"`
	ExpiresIn int64  `json:"expiresIn"`
}

type handler struct {
	serverHost string
	svc        service.URLShortener
}

func (h handler) createShortURL(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	json.NewDecoder(r.Body).Decode(&req)

	code, err := h.svc.Create(service.ShortURLInput{
		URL:       req.URL,
		ExpiresIn: req.ExpiresIn,
	})
	if err != nil {
		handleError(err, w, r)
		return
	}

	shortURL := fmt.Sprintf("%s/%s", h.serverHost, code)
	writeJSON(w, map[string]string{"url": shortURL}, http.StatusCreated)
}

func writeJSON(w http.ResponseWriter, resp interface{}, status int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func handleError(err error, w http.ResponseWriter, r *http.Request) {
	var (
		code int
		resp interface{}
	)

	if e, ok := err.(httpError); ok {
		code = e.StatusCode()
		resp = e.Response()
	} else {
		code = http.StatusInternalServerError
		resp = "internal server error"
	}

	w.Header().Add("Content-Type", jsonContentType)
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": resp,
	})
}
