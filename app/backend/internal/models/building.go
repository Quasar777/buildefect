package models

type Building struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Name    string `json:"name" gorm:"not null"`
	Address string `json:"address"`
	Stage   string `json:"stage"`
}