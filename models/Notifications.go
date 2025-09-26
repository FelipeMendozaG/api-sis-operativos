package models

import "time"

type Notification struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint       `gorm:"not null" json:"user_id"`
	Title     string     `gorm:"type:varchar(150);not null" json:"title"`
	Message   string     `gorm:"type:text;not null" json:"message"`
	Type      string     `gorm:"type:varchar(50);default:'general'" json:"type"`
	IsRead    bool       `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`

	// Relación con usuarios
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
}
