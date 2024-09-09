package routes

import (
	"new-mini-project/service1_quiz/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine)  {
	router.POST("/createquiz/:user_id",handlers.CreateQuiz)
	router.GET("/quizzes",handlers.GetAllQuizes)

	router.GET("/departments",handlers.GetDepartments)
	router.GET("/departments/:department_name",handlers.GetQuizzesByDepartmentName)

	router.GET("/quiz/:quiz_id",handlers.GetQuiz)
	router.PUT("/quiz/:quiz_id/:user_id",handlers.UpdateQuiz)
	router.DELETE("/quiz/:quiz_id/:user_id",handlers.DeleteQuiz)

	router.POST("/quiz/:quiz_id/question/:user_id",handlers.AddQuestion)
	router.GET("/quiz/:quiz_id/question/:question_id",handlers.GetQuestion)
	router.PUT("/quiz/:quiz_id/question/:question_id/:user_id",handlers.UpdateQuestion)
	router.DELETE("/quiz/:quiz_id/question/:question_id",handlers.DeleteQuestion)
}