package models

import "time"

type User struct {
	ID        uint   `gorm:"primarykey"`
	UserID    int64  `gorm:"unique;not null"`
	Username  string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	Email     string `gorm:"unique"`
	Phone     string `gorm:"unique"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
