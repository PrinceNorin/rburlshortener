package service

import (
	"testing"

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

func TestRepositoryCreateShortURL(t *testing.T) {
	suite.Run(t, new(URLShortenerRepositorySuite))
}
