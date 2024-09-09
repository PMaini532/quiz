package quizmodels

import "new-mini-project/common"


type Quiz struct {
	ID  int `json:"id" gorm:"primaryKey"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Department  string     `json:"department"`
	Questions   []Question `json:"questions" gorm:"foreignKey:QuizID;constraint:OnDelete:CASCADE;"`
}

var Quizzes []Quiz

type Department struct{
	Name string `json:"departmentname" gorm:"primaryKey"`
	NumQuizzes int `json:"noofquizzes" gorm:"default:0"`
}
var Departments []Department


type Question struct{
	ID int `json:"id" gorm:"primaryKey"`    
	QuizID int `json:"quiz_id"`
	Text string `json:"text"`
	Answer string `json:"answer"`
	Options []Option `json:"options" gorm:"foreignKey:QuestionID;constarint:OnDelete:CASCADE;"`
}

type Option struct{
	ID int `json:"id" gorm:"primaryKey"`
	Text string `json:"text"`
	IsCorrect bool `json:"is_correct"`
	QuestionID int `json:"question_id"`
}

var Questions []Question

func GetQuizzesByDepartmentName(departmentName string, quizzes *[]Quiz) error {
	// Join the Department table to filter quizzes by department name
	result := common.DB.Where("department = ?", departmentName).Find(quizzes)
	return result.Error
}