package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) CreateShortURL(shortURL *ShortURL) error {
	args := m.Called(shortURL)
	return args.Error(0)
}

func (m *mockRepo) FindShortURL(code string) (*ShortURL, error) {
	args := m.Called(code)
	if args.Get(0) != nil {
		return args.Get(0).(*ShortURL), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepo) UpdateShortURL(shortURL *ShortURL) error {
	args := m.Called(shortURL)
	return args.Error(0)
}

func (m *mockRepo) IncreaseShortURLHitCount(code string, count int) error {
	args := m.Called(code, count)
	return args.Error(0)
}

func (m *mockRepo) ListShortURLs(offset, size int64, filters ...*FilterParams) ([]*ShortURL, int64, error) {
	arguments := []interface{}{offset, size}
	for _, f := range filters {
		arguments = append(arguments, f)
	}

	args := m.Called(arguments...)
	return args.Get(0).([]*ShortURL), int64(args.Int(1)), args.Error(2)
}

func TestServiceCreateShortURL(t *testing.T) {
	type test struct {
		input ShortURLInput
		want  error
	}

	tests := []test{
		{input: ShortURLInput{URL: "example.com"}, want: ErrInvalidURL},
		{input: ShortURLInput{URL: "invalid url"}, want: ErrInvalidURL},
		{input: ShortURLInput{URL: "http://example.com", ExpiresIn: -1}, want: ErrInvalidExpiresIn},
		{input: ShortURLInput{URL: "http://example.com"}, want: nil},
	}

	repo := new(mockRepo)
	svc := NewURLShortener(repo)

	for _, tc := range tests {
		if tc.want == nil {
			repo.On("CreateShortURL", mock.Anything).
				Return(tc.want)
		}

		_, err := svc.Create(tc.input)
		if err != tc.want {
			t.Errorf("expected: %v, got: %v", tc.want, err)
		}

		repo.AssertExpectations(t)
	}
}

func TestServiceDeleteShortURL(t *testing.T) {
	type test struct {
		input string
		want  error
	}

	tests := []test{
		{
			input: "123",
			want:  nil,
		},
		{
			input: "456",
			want:  ErrRecordNotFound,
		},
	}

	repo := new(mockRepo)
	svc := NewURLShortener(repo)

	for _, tc := range tests {
		shortURL := &ShortURL{Code: tc.input}
		repo.On("FindShortURL", tc.input).
			Return(shortURL, tc.want)

		if tc.want == nil {
			repo.On("UpdateShortURL", mock.Anything).
				Return(nil)
		}

		err := svc.Delete(tc.input)
		if err != tc.want {
			t.Errorf("expected: %v, got: %v", tc.want, err)
		}
		if err == nil && (shortURL.ExpiresAt == nil || shortURL.ExpiresAt.After(time.Now().UTC())) {
			t.Error("expected to mark short url as expired")
		}

		repo.AssertExpectations(t)
	}
}

func TestServiceIncreaseShortURLHitCount(t *testing.T) {
	type test struct {
		input string
		want  error
	}

	tests := []test{
		{input: "123", want: nil},
		{input: "456", want: ErrRecordNotFound},
	}

	repo := new(mockRepo)
	svc := NewURLShortener(repo)

	for _, tc := range tests {
		repo.On("IncreaseShortURLHitCount", tc.input, 1).Return(tc.want)
		err := svc.IncreaseHitCount(tc.input)
		if err != tc.want {
			t.Errorf("expected: %v, got: %v", tc.want, err)
		}

		repo.AssertExpectations(t)
	}
}

func TestServiceGetFullURL(t *testing.T) {
	repo := new(mockRepo)

	expiresAt := time.Now().Add(-1 * time.Minute).UTC()
	s1 := &ShortURL{FullURL: "http://example.com", Code: "123"}
	s2 := &ShortURL{Code: "789", ExpiresAt: &expiresAt}
	repo.On("FindShortURL", "123").Return(s1, nil)
	repo.On("IncreaseShortURLHitCount", "123", 1).Return(nil)
	repo.On("FindShortURL", "456").Return(nil, ErrRecordNotFound)
	repo.On("FindShortURL", "789").Return(s2, nil)

	type test struct {
		input   string
		fullURL string
		err     error
	}

	tests := []test{
		{input: "123", fullURL: "http://example.com", err: nil},
		{input: "456", fullURL: "", err: ErrRecordNotFound},
		{input: "789", fullURL: "", err: ErrShortURLExpired},
	}

	svc := NewURLShortener(repo)

	for _, tc := range tests {
		fullURL, err := svc.GetFullURL(tc.input)
		if fullURL != tc.fullURL {
			t.Errorf("expected: %v, got: %v", "http://example.com", fullURL)
		}
		if err != tc.err {
			t.Errorf("expected error to be: %v, got: %v", nil, err)
		}
	}

	repo.AssertExpectations(t)
}

func TestServiceFindShortURLs(t *testing.T) {
	repo := new(mockRepo)
	svc := NewURLShortener(repo)

	var filter *FilterParams
	shortURLs := []*ShortURL{
		{FullURL: "http://example.com", Code: "123"},
		{FullURL: "http://example1.com", Code: "456"},
	}
	repo.On("ListShortURLs", int64(0), int64(30), filter).
		Return(shortURLs, 2, nil)

	r, err := svc.FindURLs(&FindParams{
		Offset: 0,
		Size:   30,
	})
	if err != nil {
		t.Errorf("expected: %v, got: %v", nil, err)
	}
	if len(r.Data) != 2 {
		t.Error("expected to get a list of short urls")
	}
	if r.TotalCount != 2 {
		t.Error("expected to get a total count of short url")
	}

	repo.AssertExpectations(t)
}
