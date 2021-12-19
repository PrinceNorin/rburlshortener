package service

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MemoryCacheStoreSuite struct {
	suite.Suite
	store CacheStore
}

func (suite *MemoryCacheStoreSuite) SetupSuite() {
	suite.store = NewMemoryCacheStore()
}

func (suite *MemoryCacheStoreSuite) TestCacheStore() {
	suite.store.Save("key1", "value1")
	suite.store.Save("key2", "value2")
	suite.store.Save("key3", "value3", CacheOption{
		ExpiresIn: 0,
	})
	suite.store.Delete("key2")

	type test struct {
		key   string
		value string
		err   error
	}

	tests := []test{
		{key: "key1", value: "value1", err: nil},
		{key: "key2", value: "", err: ErrCacheKeyNotFound},
		{key: "key3", value: "", err: ErrCacheKeyNotFound},
		{key: "key4", value: "", err: ErrCacheKeyNotFound},
	}
	for _, tc := range tests {
		var value string
		err := suite.store.Get(tc.key, &value)
		suite.Equal(tc.value, value)
		suite.Equal(tc.err, err)
	}
}

func TestMemoryCacheStore(t *testing.T) {
	suite.Run(t, new(MemoryCacheStoreSuite))
}
