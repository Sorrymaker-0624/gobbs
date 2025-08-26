package models

import "time"

type Post struct {
	ID          uint   `gorm:"primarykey"`
	AuthorID    uint   `gorm:"not null"` // [修改] 类型改为 uint
	CommunityID uint   `gorm:"not null"` // [修改] 类型改为 uint
	Status      int8   `gorm:"not null;default:1"`
	Title       string `gorm:"not null"`
	Content     string `gorm:"type:longtext;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// [修改] 简化外键关联，GORM会自动推断 AuthorID 关联 User 的主键 ID
	User User `gorm:"foreignKey:AuthorID"`
}
