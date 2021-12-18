package service

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type BlackListURLShortenerSuite struct {
	suite.Suite
	svc  URLShortener
	repo *mockRepo
}

func (suite *BlackListURLShortenerSuite) SetupSuite() {
	blacklists := []string{
		`sample.com`,
		`example\.(.+)\/block*`,
	}

	repo := new(mockRepo)
	svc, err := WithBlacklist(NewURLShortener(repo), blacklists)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.svc = svc
	suite.repo = repo
}

func (suite *BlackListURLShortenerSuite) TestCreate() {
	type test struct {
		input string
		want  error
	}

	tests := []test{
		{input: "http://example.com/123", want: nil},
		{input: "http://example.com/block/123", want: ErrBlockedURL},
		{input: "http://sample.com/123", want: ErrBlockedURL},
	}

	suite.repo.On("CreateShortURL", mock.Anything).Return(nil)
	for _, tc := range tests {
		_, err := suite.svc.Create(ShortURLInput{URL: tc.input})
		suite.Equal(tc.want, err)
	}

	suite.repo.AssertExpectations(suite.T())
}

func (suite *BlackListURLShortenerSuite) TestGetFullURL() {
	type test struct {
		input string
		want  error
	}

	tests := []test{
		{input: "123", want: nil},
		{input: "456", want: ErrBlockedURL},
		{input: "789", want: ErrBlockedURL},
	}

	suite.repo.On("FindShortURL", "123").Return(&ShortURL{
		Code:    "123",
		FullURL: "http://example.com/123",
	}, nil)
	suite.repo.On("IncreaseShortURLHitCount", "123", 1).Return(nil)
	suite.repo.On("FindShortURL", "456").Return(&ShortURL{
		Code:    "456",
		FullURL: "http://example.com/block/123",
	}, nil)
	suite.repo.On("IncreaseShortURLHitCount", "456", 1).Return(nil)
	suite.repo.On("FindShortURL", "789").Return(&ShortURL{
		Code:    "789",
		FullURL: "http://sample.com/123",
	}, nil)
	suite.repo.On("IncreaseShortURLHitCount", "789", 1).Return(nil)

	for _, tc := range tests {
		_, err := suite.svc.GetFullURL(tc.input)
		suite.Equal(tc.want, err)
	}

	suite.repo.AssertExpectations(suite.T())
}

func TestBlacklistURLShortener(t *testing.T) {
	suite.Run(t, new(BlackListURLShortenerSuite))
}
