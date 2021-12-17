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

func (suite *URLShortenerRepositorySuite) TestListShortURLs() {
	shortURLs := []ShortURL{
		{FullURL: "http://example.com", Domain: "example.com", Code: "123"},
		{FullURL: "http://testdomain.com", Domain: "testdomain.com", Code: "456"},
		{FullURL: "http://myawesome-site.com", Domain: "myawesome-site.com", Code: "789"},
	}
	for _, shortURL := range shortURLs {
		suite.repo.CreateShortURL(&shortURL)
	}

	type test struct {
		input *FindParams
		codes []string
		count int64
	}

	params := []*FindParams{
		{Offset: 0, Size: 10},
		{Offset: 1, Size: 1},
		{Offset: 0, Size: 30, Filter: &FilterParams{Code: "123"}},
		{Offset: 0, Size: 30, Filter: &FilterParams{Code: "321"}},
		{Offset: 0, Size: 30, Filter: &FilterParams{Keyword: "awesome"}},
		{Offset: 0, Size: 30, Filter: &FilterParams{Code: "123", Keyword: "awesome"}},
	}

	tests := []test{
		{input: params[0], codes: []string{"123", "456", "789"}, count: 3},
		{input: params[1], codes: []string{"456"}, count: 3},
		{input: params[2], codes: []string{"123"}, count: 1},
		{input: params[3], codes: nil, count: 0},
		{input: params[4], codes: []string{"789"}, count: 1},
		{input: params[5], codes: nil, count: 0},
	}
	for _, tc := range tests {
		shortURLs, count, _ := suite.repo.ListShortURLs(tc.input.Offset, tc.input.Size, tc.input.Filter)

		var codes []string
		for _, shortURL := range shortURLs {
			codes = append(codes, shortURL.Code)
		}
		suite.Equal(tc.count, count)
		suite.Equal(tc.codes, codes)
	}
}

func TestURLShortenerRepository(t *testing.T) {
	suite.Run(t, new(URLShortenerRepositorySuite))
}
