package transport

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/PrinceNorin/rburlshortener/service"
	"github.com/gorilla/mux"
)

var (
	jsonContentType = "application/json"
	htmlContentType = "text/html"
)

// Config server configuration
type HTTPConfig struct {
	Service    service.URLShortener
	ServerHost string
	AdminToken string
}

// NewHTTPHandler factory function
func NewHTTPHandler(conf HTTPConfig) http.Handler {
	r := mux.NewRouter()
	h := handler{svc: conf.Service, serverHost: conf.ServerHost}

	r.Use(loggingMiddleware(log.New(os.Stdout, "", 0)))
	r.Use(recoverer)
	r.HandleFunc("/", h.createShortURL).
		Methods("POST")
	r.HandleFunc("/{code}", h.getFullURL).
		Methods("GET")

	// Admin endpoints
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(httpAdminAuthMiddleware(conf.AdminToken))
	admin.HandleFunc("/shortUrls", h.adminListShortURLs).Methods("GET")
	admin.HandleFunc("/shortUrls/{code}", h.adminDeleteShortURL).Methods("DELETE")

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

func (h handler) getFullURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	fullURL, err := h.svc.GetFullURL(code)
	if err != nil {
		handleError(err, w, r)
		return
	}

	http.Redirect(w, r, fullURL, http.StatusFound)
}

func (h handler) adminListShortURLs(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.FindURLs(getFindParams(r))
	if err != nil {
		handleError(err, w, r)
		return
	}

	data := map[string]interface{}{
		"data":       result.Data,
		"totalCount": result.TotalCount,
	}
	writeJSON(w, data, http.StatusOK)
}

func (h handler) adminDeleteShortURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if err := h.svc.Delete(vars["code"]); err != nil {
		handleError(err, w, r)
		return
	}
	w.Header().Add("Content-Type", jsonContentType)
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, resp interface{}, status int) {
	w.Header().Add("Content-Type", jsonContentType)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func getFindParams(r *http.Request) *service.FindParams {
	offset, size := getPaginationParams(r)
	return &service.FindParams{
		Size:   size,
		Offset: offset,
		Filter: &service.FilterParams{
			Code:    r.URL.Query().Get("shortCode"),
			Keyword: r.URL.Query().Get("keyword"),
		},
	}
}

func getPaginationParams(r *http.Request) (int64, int64) {
	var (
		offset int64 = 0
		size   int64 = 30
	)

	val := r.URL.Query().Get("offset")
	if val != "" {
		if v, err := strconv.ParseInt(val, 10, 64); err == nil {
			offset = v
		}
	}
	val = r.URL.Query().Get("size")
	if val != "" {
		if v, err := strconv.ParseInt(val, 10, 64); err == nil {
			size = v
		}
	}

	return offset, size
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

	switch code {
	case 404, 410:
		w.Header().Add("Content-Type", htmlContentType)
		w.WriteHeader(code)
	default:
		w.Header().Add("Content-Type", jsonContentType)
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": resp,
		})
	}
}
