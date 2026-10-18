package model

import "time"

type Session struct {
	Token     string    `gorm:"primaryKey;size:64" json:"token"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
}

func (Session) TableName() string {
	return "session"
}
