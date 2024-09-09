package handlers

import (
	"net/http"
	"new-mini-project/common"
	"new-mini-project/service1_quiz/quizmodels"
	"new-mini-project/service2_user/usermodels"
	"new-mini-project/service3_test/testmodels"
	"strconv"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


func StartQuiz(c *gin.Context) {
	quizIDParam := c.Param("quiz_id")
	// userIDParam := c.Param("user_id")
	quizID, err := strconv.Atoi(quizIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Quiz Id"})
		return
	}
	// userID,err := strconv.Atoi(userIDParam)
	// if err != nil{
	// 	c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid User Id"})
	// 	return
	// }
	// var user usermodels.User
	// if err := common.DB.First(&user,userID).Error; err != nil{
	// 	c.JSON(http.StatusNotFound,gin.H{"error":"User not found"})
	// 	return
	// }
	var quiz quizmodels.Quiz
	if err := common.DB.Preload("Questions.Options").First(&quiz,quizID).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Quiz not found"})
		return
	}
	// if user.Department != quiz.Department{
	// 	c.JSON(http.StatusForbidden,gin.H{"error":"You do not belong to the department for this quiz"})
	// 	return
	// }

	var sanitizedQuestions []quizmodels.Question
    for _, question := range quiz.Questions {
        sanitizedQuestion := quizmodels.Question{
            ID:      question.ID,
            QuizID:  question.QuizID,
            Text:    question.Text,
            Options: sanitizeOptions(question.Options),
        }
        sanitizedQuestions = append(sanitizedQuestions, sanitizedQuestion)
    }
	quizData := struct {
        Quiz string         `json:"quiz"`
        Questions []quizmodels.Question `json:"questions"`
    }{
        Quiz: quiz.Title,
        Questions: sanitizedQuestions,
    }
    c.JSON(http.StatusOK, quizData)
}

func sanitizeOptions(options []quizmodels.Option) []quizmodels.Option {
	var sanitizedOptions []quizmodels.Option
	for _, option := range options {
		sanitizedOption := quizmodels.Option{
			ID:   option.ID,
			Text: option.Text,
			QuestionID: option.QuestionID,
		}
		sanitizedOptions = append(sanitizedOptions, sanitizedOption)
	}
	return sanitizedOptions
}

func SubmitQuiz(c *gin.Context) {
	quizIDParam := c.Param("quiz_id")
	userIDParam := c.Param("user_id")
	quizID, err := strconv.Atoi(quizIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Quiz Id"})
		return
	}
	userID,err := strconv.Atoi(userIDParam)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid User Id"})
		return
	}
	var user usermodels.User
	if err := common.DB.First(&user,userID).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"User not found"})
		return
	}
	var quiz quizmodels.Quiz
	if err := common.DB.Preload("Questions.Options").First(&quiz,quizID).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Quiz not found"})
		return
	}


	if user.Department != quiz.Department {
        c.JSON(http.StatusForbidden, gin.H{"error": "You do not belong to the department for this quiz"})
        return
    }

	var existingScore testmodels.QuizScore
    if err := common.DB.Where("user_id = ? AND quiz_id = ?", userID, quizID).First(&existingScore).Error; err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Quiz already submitted"})
        return
    }
	
	var answers []testmodels.Answer
	if err := c.ShouldBindJSON(&answers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var score int
	for _, answer := range answers {
        for _, question := range quiz.Questions {
            if question.ID == answer.QuestionID {
                for _, option := range question.Options {
                    if option.Text == answer.Answer {
                        if option.IsCorrect {
                            score++
                        }
                    }
                }
            }
        }
    }
	quizScore := testmodels.QuizScore{
		UserId: userID,
		QuizId: quizID,
		Score: score,
	}
	if err := common.DB.Where("user_id = ? AND quiz_id = ?",userID,quizID).FirstOrCreate(&quizScore).Error; err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save score", "details": err.Error()})
        return
	}
	c.JSON(http.StatusOK, gin.H{"score": score})
}

func GetQuizRanking(c *gin.Context)  {
	quizIDParam := c.Param("quiz_id")
	quizID,err := strconv.Atoi(quizIDParam)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid Quiz ID"})
		return
	}
	var quiz quizmodels.Quiz
	if err := common.DB.First(&quiz,quizID).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			c.JSON(http.StatusNotFound,gin.H{"error":"Quiz not found"})
		}else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quiz", "details": err.Error()})
		}
		return
	}
	var rankings []testmodels.QuizScore
	if err := common.DB.Preload("User").Where("quiz_id = ?",quizID).Order("score desc").Find(&rankings).Error; err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quiz scores", "details": err.Error()})
		return
	}
	type RankingScore struct{
		UserID int `json:"user_id"`
		Username string `json:"username"`
		Score int `json:"score"`
	}
	var RankingUsers []RankingScore
	for _,ranking := range rankings{
			RankingUsers = append(RankingUsers,RankingScore{
				UserID: ranking.UserId,
				Username: ranking.User.Name,
				Score: ranking.Score,
			})
	}
	c.JSON(http.StatusOK,RankingUsers)
}

func GetUserScores(c *gin.Context)  {
	userIDParam := c.Param("user_id")
    userID, err := strconv.Atoi(userIDParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User Id"})
        return
    }

    var user usermodels.User
    if err := common.DB.First(&user, userID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
	var scores []testmodels.QuizScore
    if err := common.DB.Where("user_id = ?", userID).Preload("Quiz").Find(&scores).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve scores", "details": err.Error()})
        return
    }
	var response []map[string]interface{}
    for _, score := range scores {
        response = append(response, map[string]interface{}{
            "quiz_id":   score.QuizId,
            "quiz_name": score.Quiz.Title,
            "score":     score.Score,
        })
    }

    c.JSON(http.StatusOK, response)	
}
