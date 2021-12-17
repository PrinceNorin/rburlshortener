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
