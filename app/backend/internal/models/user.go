package models

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Login    string `json:"login" gorm:"size:100;not null;unique"`
	PasswordHash string `json:"password" gorm:"not null"`
	Name     string `json:"name" gorm:"not null"`
	LastName string `json:"lastname" gorm:"not null"`
	Role     string `json:"role" gorm:"not null"` // engineer, manager, observer
}








