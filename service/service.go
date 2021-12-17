package service

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"time"
)

// Max random short code length
const MAX_SHORT_CODE_LENGTH = 12

// Errors return from service
var (
	ErrInvalidURL       = newError("invalid url", http.StatusBadRequest)
	ErrInvalidExpiresIn = newError("invalid expires in", http.StatusBadRequest)
	ErrRecordNotFound   = newError("record not found", http.StatusNotFound)
	ErrShortURLExpired  = newError("url expired", http.StatusGone)
)

// ShortURLInput used to create a ShortURL
type ShortURLInput struct {
	URL       string
	ExpiresIn int64 // second
}

// FindParams used to get/filter short urls
type FindParams struct {
	Offset int64
	Size   int64
	Filter *FilterParams
}

// FilterParams input to filter ShortURL
type FilterParams struct {
	Code    string
	Keyword string
}

// Result type returned by FindURLs
type Result struct {
	Data       []*ShortURL
	TotalCount int64
}

// URLShortener public service interface
type URLShortener interface {
	// Create a new short url
	Create(input ShortURLInput) (string, error)
	// FindURLs return a list of short urls
	FindURLs(params *FindParams) (*Result, error)
	// Delete a short url
	Delete(code string) error
	// IncreaseHitCount of a short url
	IncreaseHitCount(code string) error
	// Get full url from code and increase hit count
	GetFullURL(code string) (string, error)
}

// NewURLShortener factory function
func NewURLShortener(repo URLShortenerRepository) URLShortener {
	return &urlShortener{repo: repo}
}

type urlShortener struct {
	repo URLShortenerRepository
}

func (s *urlShortener) Create(input ShortURLInput) (string, error) {
	u, err := url.ParseRequestURI(input.URL)
	if err != nil || u.Host == "" {
		return "", ErrInvalidURL
	}
	domain := u.Host
	if u.Port() != "" {
		domain += ":" + u.Port()
	}

	code, err := getRandomShortCode(MAX_SHORT_CODE_LENGTH)
	if err != nil {
		return "", err
	}

	shortURL := ShortURL{
		Code:    code,
		FullURL: input.URL,
		Domain:  domain,
	}
	if input.ExpiresIn < 0 {
		return "", ErrInvalidExpiresIn
	}
	if input.ExpiresIn > 0 {
		d := time.Duration(input.ExpiresIn) * time.Second
		expiresAt := time.Now().Add(d).UTC()
		shortURL.ExpiresAt = &expiresAt
	}

	if err := s.repo.CreateShortURL(&shortURL); err != nil {
		return "", err
	}
	return code, nil
}

func (s *urlShortener) FindURLs(params *FindParams) (*Result, error) {
	shortURLs, count, err := s.repo.ListShortURLs(params.Offset, params.Size, params.Filter)
	if err != nil {
		return nil, err
	}

	return &Result{
		Data:       shortURLs,
		TotalCount: count,
	}, nil
}

func (s *urlShortener) Delete(code string) error {
	shortURL, err := s.repo.FindShortURL(code)
	if err != nil {
		return ErrRecordNotFound
	}

	// mark this short url as expired
	expiresAt := time.Now().Add(-1 * time.Hour).UTC()
	shortURL.ExpiresAt = &expiresAt
	if err := s.repo.UpdateShortURL(shortURL); err != nil {
		return err
	}

	return nil
}

func (s *urlShortener) IncreaseHitCount(code string) error {
	return s.repo.IncreaseShortURLHitCount(code, 1)
}

func (s *urlShortener) GetFullURL(code string) (string, error) {
	shortURL, err := s.repo.FindShortURL(code)
	if err != nil {
		return "", ErrRecordNotFound
	}
	// check if short url is expired
	if shortURL.ExpiresAt != nil && shortURL.ExpiresAt.Before(time.Now().UTC()) {
		return "", ErrShortURLExpired
	}
	if err := s.repo.IncreaseShortURLHitCount(shortURL.Code, 1); err != nil {
		return "", err
	}
	return shortURL.FullURL, nil
}

func getRandomShortCode(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
