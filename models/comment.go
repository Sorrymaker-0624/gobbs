package models

import "time"

type Comment struct {
	ID        uint   `gorm:"primarykey"`
	CommentID int64  `gorm:"unique;not null"`
	PostID    int64  `gorm:"not null"`
	AuthorID  int64  `gorm:"not null"`
	Content   string `gorm:"type:text;not null"`
	Status    int8   `gorm:"not null;default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
