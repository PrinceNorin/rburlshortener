package service

import (
	"gorm.io/gorm"
)

// ShortURLInput used to create a ShortURL
type ShortURLInput struct {
	URL       string
	ExpiresAt string
}

// FindParams used to get/filter short urls
type FindParams struct {
	Offset int64
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
}

// NewURLShortener factory function
func NewURLShortener(db *gorm.DB) URLShortener {
	return &urlShortener{db: db}
}

type urlShortener struct {
	db *gorm.DB
}

func (s *urlShortener) Create(input ShortURLInput) (string, error) {
	return "", nil
}

func (s *urlShortener) FindURLs(params *FindParams) (*Result, error) {
	return nil, nil
}

func (s *urlShortener) Delete(code string) error {
	return nil
}

func (s *urlShortener) IncreaseHitCount(code string) error {
	return nil
}
