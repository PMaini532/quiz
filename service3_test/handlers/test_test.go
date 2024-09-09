package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"new-mini-project/common"
	"new-mini-project/service1_quiz/quizmodels"
	"new-mini-project/service2_user/usermodels"
	"new-mini-project/service3_test/testmodels"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB() (*gorm.DB, error) {
	dsn := "host=localhost user=maini password=pratham dbname=quizdb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&quizmodels.Quiz{}, &quizmodels.Department{}, &usermodels.User{},&usermodels.DeletedID{},&quizmodels.Question{},&quizmodels.Option{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/quizzes/:quiz_id/start/:user_id", StartQuiz)
	r.POST("/quizzes/:quiz_id/submit/:user_id", SubmitQuiz)
	r.GET("/quiz/:quiz_id/ranking",GetQuizRanking)
	r.GET("/users/scores/:user_id",GetUserScores)
	return r
}

func TestStartAndSubmitQuiz(t *testing.T) {
    db, err := setupTestDB()
    if err != nil {
        t.Fatalf("Failed to set up test database: %v", err)
    }
    common.DB = db
    router := setupRouter()

    userID := 4 
    quizID := 1 


    startQuizReq, _ := http.NewRequest("POST", fmt.Sprintf("/quizzes/%d/start/%d", quizID, userID), nil)
    startQuizW := httptest.NewRecorder()
    router.ServeHTTP(startQuizW, startQuizReq)

    assert.Equal(t, http.StatusOK, startQuizW.Code)

    var startQuizResponse struct {
        Quiz       string                   `json:"quiz"`
        Questions  []quizmodels.Question    `json:"questions"`
    }
    err = json.Unmarshal(startQuizW.Body.Bytes(), &startQuizResponse)
    assert.NoError(t, err)
    assert.Equal(t, "Updated Quiz Title", startQuizResponse.Quiz)
    assert.NotEmpty(t, startQuizResponse.Questions)

 
    var questionIDs []int
    for _, question := range startQuizResponse.Questions {
        assert.NotEmpty(t, question.Text)
        assert.NotEmpty(t, question.Options)
        questionIDs = append(questionIDs, question.ID)
    }


    var answers []testmodels.Answer
    for _, questionID := range questionIDs {
        
        answers = append(answers, testmodels.Answer{
            QuestionID: questionID,
            Answer:     "4", 
        })
    }
    answersJSON, _ := json.Marshal(answers)

    submitQuizReq, _ := http.NewRequest("POST", fmt.Sprintf("/quizzes/%d/submit/%d", quizID, userID), bytes.NewBuffer(answersJSON))
    submitQuizReq.Header.Set("Content-Type", "application/json")
    submitQuizW := httptest.NewRecorder()
    router.ServeHTTP(submitQuizW, submitQuizReq)

    assert.Equal(t, http.StatusOK, submitQuizW.Code)

    var submitQuizResponse map[string]interface{}
    err = json.Unmarshal(submitQuizW.Body.Bytes(), &submitQuizResponse)
    assert.NoError(t, err)
    expectedScore := calculateExpectedScore(answers) 
    assert.Equal(t, float64(expectedScore), submitQuizResponse["score"])
}


func calculateExpectedScore(answers []testmodels.Answer) int {
    var score int
    for _, answer := range answers {
        var question quizmodels.Question
        err := common.DB.Preload("Options").First(&question, answer.QuestionID).Error
        if err != nil {
            continue 
        }
        for _, option := range question.Options {
            if option.Text == answer.Answer && option.IsCorrect {
                score++
            }
        }
    }
    return score
}

func TestGetUserScores(t *testing.T)  {
    db,err := setupTestDB()
    if err != nil{
        t.Fatalf("Fail to set up database: %v", err)
    }
    common.DB = db
    router := setupRouter()
    userID := 4
    expectedScores := []map[string]interface{}{
        {
            "quiz_id" : 1.0,
            "quiz_name" : "Updated Quiz Title",
            "score" : 2.0,
        },
    }
    req,_ := http.NewRequest("GET",fmt.Sprintf("/users/scores/%d",userID),nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w,req)
    
    assert.Equal(t,http.StatusOK,w.Code)
    var response []map[string]interface{}
    err = json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, len(expectedScores), len(response))
    for i, score := range response {
        assert.Equal(t, expectedScores[i]["quiz_id"], score["quiz_id"])
        assert.Equal(t, expectedScores[i]["quiz_name"], score["quiz_name"])
        assert.Equal(t, expectedScores[i]["score"], score["score"])
    }

}