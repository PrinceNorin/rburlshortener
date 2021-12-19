package service

import "time"

// ShortURL model mapping to short_urls table
type ShortURL struct {
	Id        int64      `json:"-"`
	FullURL   string     `json:"fullUrl" gorm:"not null"`
	Domain    string     `json:"-" gorm:"not null;index"`
	Code      string     `json:"code" gorm:"unique;not null"`
	HitCount  int64      `json:"hitCount" gorm:"default:0"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	CreatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-" gorm:"index"`
}
