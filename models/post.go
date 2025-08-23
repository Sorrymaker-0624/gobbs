package models

import "time"

type Post struct {
	ID          uint   `gorm:"primarykey"`
	PostID      int64  `gorm:"unique;not null"`
	AuthorID    int64  `gorm:"not null"`
	CommunityID int64  `gorm:"not null"`
	Status      int8   `gorm:"not null;default:1"`
	Title       string `gorm:"not null"`
	Content     string `gorm:"type:longtext;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
