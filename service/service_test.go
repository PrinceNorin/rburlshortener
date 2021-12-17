package service

import "testing"

type mockRepo struct {
	shortURLS []*ShortURL
}

func (m *mockRepo) CreateShortURL(shortURL *ShortURL) error {
	m.shortURLS = append(m.shortURLS, shortURL)
	return nil
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

	repo := &mockRepo{}
	svc := NewURLShortener(repo)

	for _, tc := range tests {
		_, err := svc.Create(tc.input)
		if err != tc.want {
			t.Fatalf("expected: %v, got: %v", tc.want, err)
		}
	}

	if len(repo.shortURLS) == 0 {
		t.Error("error failed to create short url")
	}
}
