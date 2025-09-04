package models

import (
	"time"

	"gorm.io/gorm"
)

type Attendance struct {
	gorm.Model
	UserID uint      `json:"user_id"`
	Date   time.Time `json:"date" gorm:"type:timestamp;not null"`
	Status string    `json:"status"`
}
