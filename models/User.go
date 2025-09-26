package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName    string `json:"first_name" gorm:"type:varchar(255);not null"`
	LastName     string `json:"last_name" gorm:"type:varchar(255);not null"`
	Email        string `json:"email" gorm:"unique_index;type:varchar(255);not null"`
	Password     string `json:"password" gorm:"not null"`
	PasswordHash string `json:"password_hash"`
	RoleID       uint   `json:"role_id"`
	Role         Role   `json:"role"`
	Status       bool   `json:"status" gorm:"type:boolean"`
}
