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
	FindShortURL(code string) (*ShortURL, error)
	UpdateShortURL(shortURL *ShortURL) error
	IncreaseShortURLHitCount(code string, count int) error
	ListShortURLs(offset, size int64, filters ...*FilterParams) ([]*ShortURL, int64, error)
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

func (r *sqliteRepository) FindShortURL(code string) (*ShortURL, error) {
	var shortURL ShortURL
	if err := r.db.Unscoped().Where("code = ?", code).First(&shortURL).Error; err != nil {
		return nil, err
	}
	return &shortURL, nil
}

func (r *sqliteRepository) UpdateShortURL(shortURL *ShortURL) error {
	if shortURL.Id == 0 {
		return ErrRecordNotFound
	}
	return r.db.Model(&ShortURL{Id: shortURL.Id}).Updates(shortURL).Error
}

func (r *sqliteRepository) IncreaseShortURLHitCount(code string, count int) error {
	result := r.db.Model(&ShortURL{}).Where("code = ?", code).
		Update("hit_count", gorm.Expr("hit_count + ?", count))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (r *sqliteRepository) ListShortURLs(offset, size int64, filters ...*FilterParams) ([]*ShortURL, int64, error) {
	var filter *FilterParams
	if len(filters) > 0 {
		filter = filters[0]
	}
	if offset < 0 {
		offset = 0
	}

	scope := r.db.Model(&ShortURL{})
	if filter != nil {
		if filter.Code != "" {
			scope = scope.Where("code = ?", filter.Code)
		}
		if filter.Keyword != "" {
			scope = scope.Where("domain LIKE ?", "%"+filter.Keyword+"%")
		}
	}

	var count int64
	var shortURLs []*ShortURL

	if err := scope.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	scope = scope.Offset(int(offset)).Limit(int(size))
	if err := scope.Find(&shortURLs).Error; err != nil {
		return nil, 0, err
	}

	return shortURLs, count, nil
}

func transformError(err error) error {
	if e, ok := err.(sqlite3.Error); ok && e.Code == 19 {
		return ErrConstraintUnique
	}
	return err
}
