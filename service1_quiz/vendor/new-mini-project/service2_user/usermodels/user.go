package usermodels

type User struct{
	ID int `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
	Email string `json:"email" gorm:"unique"`
	Password string `json:"password"`
	Department string `json:"department,omitempty"`
	IsAdmin bool `json:"is_admin"`
}

type DeletedID struct{
	ID int `gorm:"primaryKey"`
}
var Users []User
