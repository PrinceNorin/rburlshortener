package service

import "time"

// ShortURL model mapping to short_urls table
type ShortURL struct {
	Id        int64      `json:"id"`
	FullURL   string     `json:"fullUrl" gorm:"not null"`
	Domain    string     `json:"domain" gorm:"not null;index"`
	Code      string     `json:"code" gorm:"unique;not null"`
	HitCount  int64      `json:"hitCount" gorm:"default:0"`
	ExpiresAt *time.Time `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
}
