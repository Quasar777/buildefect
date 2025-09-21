package models

import "time"

type Defect struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	BuildingID          uint      `json:"building_id"`
	Building            Building  `json:"building" gorm:"foreignKey:BuildingID"`
	CreatedAt           time.Time `json:"created_at"`
	CreatedByPersonID   uint      `json:"created_by_person_id"`
	CreatedBy           User      `json:"created_by" gorm:"foreignKey:CreatedByPersonID"`
	UpdatedAt           time.Time `json:"updated_at"`
	UpdatedByPersonID   uint      `json:"updated_by_person_id"`
	UpdatedBy           User      `json:"updated_by" gorm:"foreignKey:UpdatedByPersonID"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	Priority            string    `json:"priority"`   // low, medium, high
	ResponsiblePersonID uint      `json:"responsible_person_id"`
	Responsible         User      `json:"responsible" gorm:"foreignKey:ResponsiblePersonID"`
	Deadline            time.Time `json:"deadline"`
	Status              string    `json:"status"`     // new, in_progress, review, closed, cancelled
	Attachments         []DefectAttachment `json:"attachments"`
	Comments            []Comment          `json:"comments"`
}