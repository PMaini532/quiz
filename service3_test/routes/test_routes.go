package routes

import (
	"new-mini-project/service3_test/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine)  {
	router.GET("/quizzes/:quiz_id/start", handlers.StartQuiz)
	router.POST("/quizzes/:quiz_id/submit/:user_id", handlers.SubmitQuiz)
	router.GET("/quiz/:quiz_id/ranking",handlers.GetQuizRanking)
	router.GET("/users/scores/:user_id",handlers.GetUserScores)
}