package service

import (
	"time"
)

// WithCache decorate existing URLShortenerRepository with caching capability
func WithCache(repo URLShortenerRepository, store CacheStore) URLShortenerRepository {
	return &cacheRepository{
		URLShortenerRepository: repo,
		store:                  store,
	}
}

type cacheRepository struct {
	URLShortenerRepository
	store CacheStore
}

// This method implemented a simple expiring cache mechanism for demonstration purposes
func (s *cacheRepository) FindShortURL(code string) (*ShortURL, error) {
	if cache := s.getCache(code); cache != nil {
		return cache, nil
	}

	shortURL, err := s.URLShortenerRepository.FindShortURL(code)
	if err != nil {
		return nil, err
	}
	err = s.store.Save(shortURL.Code, shortURL, CacheOption{
		ExpiresIn: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return shortURL, nil
}

// Cache busting on update
func (s *cacheRepository) UpdateShortURL(shortURL *ShortURL) error {
	s.store.Delete(shortURL.Code)
	return s.URLShortenerRepository.UpdateShortURL(shortURL)
}

func (s *cacheRepository) getCache(key string) *ShortURL {
	var shortURL ShortURL
	if err := s.store.Get(key, &shortURL); err != nil {
		return nil
	}
	return &shortURL
}
