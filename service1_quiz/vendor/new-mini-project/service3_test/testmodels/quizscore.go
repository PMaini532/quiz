package testmodels
import "new-mini-project/service1_quiz/quizmodels"
import "new-mini-project/service2_user/usermodels"

type QuizScore struct{
	UserId int `gorm:"not null"`
	QuizId int `gorm:"not null"`
	Score int  `gorm:"not null"`
	User usermodels.User
	Quiz quizmodels.Quiz
}

type Answer struct {
	QuestionID int    `json:"question_id"`
	Answer     string `json:"answer"`
}