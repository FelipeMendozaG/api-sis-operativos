package models

type Permission struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Action string `gorm:"unique;not null" json:"action"`
	Roles  []Role `gorm:"many2many:role_permissions;" json:"-"`
}
