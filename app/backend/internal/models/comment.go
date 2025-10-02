package models

import "time"

type Comment struct {
	ID                uint              `json:"id" gorm:"primaryKey"`
	DefectID          uint              `json:"defect_id"`
	Defect            Defect            `json:"-" gorm:"foreignKey:DefectID"`
	CreatedAt         time.Time         `json:"created_at"`
	CreatedByPersonID uint              `json:"created_by_person_id"`
	CreatedBy         User              `json:"created_by" gorm:"foreignKey:CreatedByPersonID"`
	Text              string            `json:"text"`
	Attachments       []CommentAttachment `json:"attachments"`
}