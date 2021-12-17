package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	testDBName = "../test.sqlite"

	repo URLShortenerRepository
)

type URLShortenerRepositorySuite struct {
	suite.Suite
	db   *gorm.DB
	repo URLShortenerRepository
}

func (suite *URLShortenerRepositorySuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(testDBName))
	if err != nil {
		panic(err)
	}

	suite.db = db
	suite.repo = NewURLShortenerRepository(db)
}

func (suite *URLShortenerRepositorySuite) SetupTest() {
	suite.db.AutoMigrate(&ShortURL{})
}

func (suite *URLShortenerRepositorySuite) TearDownTest() {
	suite.db.Exec("DROP TABLE short_urls")
}

func (suite *URLShortenerRepositorySuite) TearDownSuite() {
	db, err := suite.db.DB()
	if err != nil {
		panic(err)
	}
	db.Close()
}

func (suite *URLShortenerRepositorySuite) TestCreateShortURL() {
	type test struct {
		input *ShortURL
		want  error
	}

	tests := []test{
		{input: &ShortURL{FullURL: "http://example.com", Domain: "example.com", Code: "123"}, want: nil},
		{input: &ShortURL{FullURL: "http://example1.com", Domain: "example1.com", Code: "456"}, want: nil},
		{input: &ShortURL{FullURL: "http://example2.com", Domain: "example2.com", Code: "456"}, want: ErrConstraintUnique},
	}

	for _, tc := range tests {
		err := suite.repo.CreateShortURL(tc.input)
		suite.Equal(tc.want, err)
	}
}

func (suite *URLShortenerRepositorySuite) TestFindShortURL() {
	shortURLs := []*ShortURL{
		{FullURL: "http://example.com", Domain: "example.com", Code: "123"},
		{FullURL: "http://example1.com", Domain: "example1.com", Code: "456"},
		{FullURL: "http://example2.com", Domain: "example2.com", Code: "789"},
	}
	for _, s := range shortURLs {
		suite.repo.CreateShortURL(s)
	}

	type test struct {
		input    string
		shortURL *ShortURL
		want     error
	}

	tests := []test{
		{input: "123", want: nil, shortURL: shortURLs[0]},
		{input: "321", want: gorm.ErrRecordNotFound, shortURL: nil},
	}
	for _, tc := range tests {
		s, err := suite.repo.FindShortURL(tc.input)
		suite.Equal(tc.want, err)

		if s != nil {
			suite.Equal(tc.shortURL.Id, s.Id)
		}
	}
}

func (suite *URLShortenerRepositorySuite) TestUpdateShortURL() {
	shortURL := ShortURL{FullURL: "http://example.com", Domain: "example.com", Code: "123"}
	suite.repo.CreateShortURL(&shortURL)

	type test struct {
		input *ShortURL
		want  error
	}

	expiresAt := time.Now().Add(10 * time.Minute).UTC()
	tests := []test{
		{
			input: &ShortURL{
				Id:        shortURL.Id,
				ExpiresAt: &expiresAt,
			},
			want: nil,
		},
		{
			input: &ShortURL{
				FullURL: "http://example1.com",
			},
			want: ErrRecordNotFound,
		},
	}

	for _, tc := range tests {
		err := suite.repo.UpdateShortURL(tc.input)
		suite.Equal(tc.want, err)
	}
}

func (suite *URLShortenerRepositorySuite) TestIncreaseShortURLHitCount() {
	suite.repo.CreateShortURL(&ShortURL{FullURL: "http://example.com", Domain: "example.com", Code: "123"})

	err := suite.repo.IncreaseShortURLHitCount("123", 1)
	suite.Equal(nil, err)
	s1, _ := suite.repo.FindShortURL("123")
	suite.EqualValues(1, s1.HitCount)

	err = suite.repo.IncreaseShortURLHitCount("456", 1)
	suite.Equal(ErrRecordNotFound, err)
}

func TestRepositoryCreateShortURL(t *testing.T) {
	suite.Run(t, new(URLShortenerRepositorySuite))
}
