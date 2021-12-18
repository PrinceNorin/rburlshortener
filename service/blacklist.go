package service

import "regexp"

func WithBlacklist(svc URLShortener, patterns []string) (URLShortener, error) {
	rxPatterns := []*regexp.Regexp{}
	for _, pattern := range patterns {
		r, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		rxPatterns = append(rxPatterns, r)
	}

	return &blacklistUrlShortener{
		URLShortener: svc,
		patterns:     rxPatterns,
	}, nil
}

type blacklistUrlShortener struct {
	URLShortener
	patterns []*regexp.Regexp
}

func (s *blacklistUrlShortener) Create(input ShortURLInput) (string, error) {
	if err := s.validate(input.URL); err != nil {
		return "", err
	}
	return s.URLShortener.Create(input)
}

func (s *blacklistUrlShortener) GetFullURL(code string) (string, error) {
	fullURL, err := s.URLShortener.GetFullURL(code)
	if err != nil {
		return "", err
	}
	if err := s.validate(fullURL); err != nil {
		return "", err
	}
	return fullURL, nil
}

func (s *blacklistUrlShortener) validate(url string) error {
	for _, pattern := range s.patterns {
		if pattern.MatchString(url) {
			return ErrBlockedURL
		}
	}
	return nil
}
