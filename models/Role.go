package models

type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"unique;not null" json:"name"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
	Users       []User       `json:"-"`
}
