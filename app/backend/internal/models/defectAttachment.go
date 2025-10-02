package models

type DefectAttachment struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	DefectID uint   `json:"defect_id"`
	Defect   Defect `json:"-" gorm:"foreignKey:DefectID"`
	URL      string `json:"url"`
}