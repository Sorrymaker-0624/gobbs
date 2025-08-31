package models

import "time"

type Comment struct {
	ID        uint   `gorm:"primarykey"`
	PostID    uint   `gorm:"not null"` // [修改] 类型改为 uint
	AuthorID  uint   `gorm:"not null"` // [修改] 类型改为 uint
	Content   string `gorm:"type:text;not null"`
	Status    int8   `gorm:"not null;default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
	User      User `gorm:"foreignKey:AuthorID"`
}
