package service

import (
	"errors"

	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

var (
	ErrConstraintUnique = errors.New("unique constraint failed")
)

// URLShortenerRepository to interact with data store
type URLShortenerRepository interface {
	CreateShortURL(shortURL *ShortURL) error
}

// NewURLShortenerRepository factory function
func NewURLShortenerRepository(db *gorm.DB) URLShortenerRepository {
	return &sqliteRepository{db: db}
}

type sqliteRepository struct {
	db *gorm.DB
}

func (r *sqliteRepository) CreateShortURL(shortURL *ShortURL) error {
	if err := r.db.Save(shortURL).Error; err != nil {
		return transformError(err)
	}
	return nil
}

func transformError(err error) error {
	if e, ok := err.(sqlite3.Error); ok && e.Code == 19 {
		return ErrConstraintUnique
	}
	return err
}
