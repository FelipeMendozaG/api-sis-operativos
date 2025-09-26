package models

import "time"

type RoleSchedule struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleID    uint      `gorm:"not null" json:"role_id"`
	DayOfWeek string    `gorm:"type:varchar(20);not null" json:"day_of_week"`
	StartTime string    `gorm:"type:time;not null" json:"start_time"`
	EndTime   string    `gorm:"type:time;not null" json:"end_time"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relación con Role
	Role Role `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE" json:"role"`
}
