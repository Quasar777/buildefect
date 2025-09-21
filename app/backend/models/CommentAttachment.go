package models

type CommentAttachment struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	CommentID uint    `json:"comment_id"`
	Comment   Comment `json:"-" gorm:"foreignKey:CommentID"`
	URL       string  `json:"url"`
}