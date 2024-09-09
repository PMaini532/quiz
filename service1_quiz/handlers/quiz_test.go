package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"new-mini-project/common"
	"new-mini-project/service1_quiz/quizmodels"
	"new-mini-project/service2_user/usermodels"

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
	r.POST("/createquiz/:user_id", CreateQuiz)
	r.GET("/quizzes", GetAllQuizes)
	r.GET("/quiz/:quiz_id", GetQuiz)
	r.PUT("/quiz/:quiz_id/:user_id", UpdateQuiz)
	r.POST("/quiz/:quiz_id/question/:user_id",AddQuestion)
	r.GET("/quiz/:quiz_id/question/:question_id",GetQuestion)
	r.PUT("/quiz/:quiz_id/question/:question_id/:user_id",UpdateQuestion)
	r.DELETE("/quiz/:quiz_id/question/:question_id",DeleteQuestion)
	r.DELETE("/quiz/:quiz_id/:user_id", DeleteQuiz)
	return r
}

func TestCreateQuiz(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	common.DB = db
	router := setupRouter()
	quiz := quizmodels.Quiz{
		Title:       "Test Quiz",
		Description: "A quiz for testing",
		Department:  "Science",
	}
	quizJSON, _ := json.Marshal(quiz)

	req, _ := http.NewRequest("POST", "/createquiz/1", bytes.NewBuffer(quizJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var createdQuiz quizmodels.Quiz
	err = json.Unmarshal(w.Body.Bytes(), &createdQuiz)
	assert.NoError(t, err)
	assert.Equal(t, quiz.Title, createdQuiz.Title)
	assert.Equal(t, quiz.Description, createdQuiz.Description)
	assert.Equal(t, quiz.Department, createdQuiz.Department)
}

func TestGetAllQuizes(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	common.DB = db

	router := setupRouter()

	req, _ := http.NewRequest("GET", "/quizzes", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var quizzes []quizmodels.Quiz
	err = json.Unmarshal(w.Body.Bytes(), &quizzes)
	assert.NoError(t, err)
	assert.Len(t, quizzes, 1)
}
func TestGetQuiz(t *testing.T) {

    db, err := setupTestDB()
    if err != nil {
        t.Fatalf("Failed to set up test database: %v", err)
    }
    common.DB = db

    router := setupRouter()


    quizID := 1 
    url := fmt.Sprintf("/quiz/%d", quizID)

    req, _ := http.NewRequest("GET", url, nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)


    var retrievedQuiz quizmodels.Quiz
    err = json.Unmarshal(w.Body.Bytes(), &retrievedQuiz)
    if err != nil {
        t.Fatalf("Failed to parse response body: %v", err)
    }
    var expectedQuiz quizmodels.Quiz
    if err := common.DB.First(&expectedQuiz, quizID).Error; err != nil {
        t.Fatalf("Failed to retrieve quiz from the database: %v", err)
    }

    assert.Equal(t, expectedQuiz.Title, retrievedQuiz.Title)
    assert.Equal(t, expectedQuiz.Description, retrievedQuiz.Description)
    assert.Equal(t, expectedQuiz.Department, retrievedQuiz.Department)
}

func TestUpdateQuiz(t *testing.T)  {
	db, err := setupTestDB()
    if err != nil {
        t.Fatalf("Failed to set up test database: %v", err)
    }
    common.DB = db
	router := setupRouter()
	quizID := 1
	 updatedQuiz := quizmodels.Quiz{
        Title:       "Updated Quiz Title",
        Description: "Updated description",
    }
    updatedQuizJSON, _ := json.Marshal(updatedQuiz)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/quiz/%d/1", quizID), bytes.NewBuffer(updatedQuizJSON))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var updatedQuizResponse quizmodels.Quiz
    err = json.Unmarshal(w.Body.Bytes(), &updatedQuizResponse)
    if err != nil {
        t.Fatalf("Failed to parse response body: %v", err)
    }
	var quizFromDB quizmodels.Quiz
    if err := common.DB.First(&quizFromDB, quizID).Error; err != nil {
        t.Fatalf("Failed to retrieve quiz from the database: %v", err)
    }
	assert.Equal(t, updatedQuiz.Title, quizFromDB.Title)
    assert.Equal(t, updatedQuiz.Description, quizFromDB.Description)
    assert.Equal(t, quizID, quizFromDB.ID)
}
var createdQuestionID int

func TestAddQuestion(t *testing.T)  {
	db,err := setupTestDB()
	if err != nil{
		t.Fatalf("Failed to set up databse: %v",err)
	}
	common.DB = db
	router := setupRouter()
	quizID := 1
	userID := 1
	question := quizmodels.Question{
		Text: "What is 2+2",
		QuizID: quizID,
		Answer: "4",
		Options: []quizmodels.Option{
			{Text: "4",IsCorrect: true},
			{Text: "3", IsCorrect: false},
			{Text: "5",IsCorrect: false},
			{Text: "6",IsCorrect: false},
		},
	}
	questionJSON,_ := json.Marshal(question)
	req,_ := http.NewRequest("POST",fmt.Sprintf("/quiz/%d/question/%d", quizID, userID),bytes.NewBuffer(questionJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var createdQuestion quizmodels.Question
	err = json.Unmarshal(w.Body.Bytes(), &createdQuestion)
	assert.NoError(t, err)
	assert.Equal(t, question.Text, createdQuestion.Text)
	assert.Equal(t, question.QuizID, createdQuestion.QuizID)
	assert.Len(t, createdQuestion.Options, 4)

	createdQuestionID = createdQuestion.ID
}

func TestUpdateQuestion(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up database: %v", err)
	}
	common.DB = db
	router := setupRouter()
	quizID := 1
	userID := 1
	question := quizmodels.Question{
		Text:    "What is 2+2",
		Answer: "4",
		QuizID:  quizID,
		Options: []quizmodels.Option{
			{Text: "4", IsCorrect: true},
			{Text: "3", IsCorrect: false},
			{Text: "5", IsCorrect: false},
			{Text: "6", IsCorrect: false},
		},
	}
	questionJSON, _ := json.Marshal(question)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/quiz/%d/question/%d", quizID, userID), bytes.NewBuffer(questionJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var createdQuestion quizmodels.Question
	err = json.Unmarshal(w.Body.Bytes(), &createdQuestion)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	createdQuestionID := createdQuestion.ID
	updatedQuestion := quizmodels.Question{
		Text:    "What is 1+3",
		Answer: "4",
	}
	updatedQuestionJSON, _ := json.Marshal(updatedQuestion)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/quiz/%d/question/%d/%d", quizID, createdQuestionID, userID), bytes.NewBuffer(updatedQuestionJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var updatedQuestionResponse quizmodels.Question
	err = json.Unmarshal(w.Body.Bytes(), &updatedQuestionResponse)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	assert.Equal(t, updatedQuestion.Text, updatedQuestionResponse.Text)
	assert.Equal(t, quizID, updatedQuestionResponse.QuizID)
	assert.Len(t, updatedQuestionResponse.Options, 4)
	var questionFromDB quizmodels.Question
	if err := common.DB.Preload("Options").First(&questionFromDB, createdQuestionID).Error; err != nil {
        t.Fatalf("Failed to retrieve updated question from the database: %v", err)
    }

	assert.Equal(t, updatedQuestion.Text, questionFromDB.Text)
	assert.Equal(t, quizID, questionFromDB.QuizID)
	assert.Len(t, questionFromDB.Options, 4)
}

func TestDeleteQuestion(t *testing.T)  {
	db,err := setupTestDB()
	if err != nil{
		t.Fatalf("Failed to set up database: %v", err)
	}
	common.DB=db
	router := setupRouter()
	quizID := 1
    userID := 1
	question := quizmodels.Question{
        Text:    "What is 2+2",
		Answer: "4",
        QuizID:  quizID,
        Options: []quizmodels.Option{
            {Text: "3", IsCorrect: false},
            {Text: "4", IsCorrect: true},
            {Text: "5", IsCorrect: false},
            {Text: "6", IsCorrect: false},
        },
    }
	questionJSON,_ := json.Marshal(question)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/quiz/%d/question/%d", quizID, userID), bytes.NewBuffer(questionJSON))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
	 var createdQuestion quizmodels.Question
    err = json.Unmarshal(w.Body.Bytes(), &createdQuestion)
    if err != nil {
        t.Fatalf("Failed to parse response body: %v", err)
    }
	createdQuestionID := createdQuestion.ID

	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/quiz/%d/question/%d", quizID, createdQuestionID), nil)
    w = httptest.NewRecorder()
    router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var questionFromDB quizmodels.Question
    err = common.DB.First(&questionFromDB, createdQuestionID).Error
    if err == nil {
        t.Fatalf("Expected question to be deleted, but it was found in the database")
    } else if err != gorm.ErrRecordNotFound {
        t.Fatalf("Failed to check if question was deleted: %v", err)
    }
}


func TestDeleteQuiz(t *testing.T){
	db, err := setupTestDB()
	if err != nil{
		t.Fatalf("Failed to set up database: %v", err)
	}
	common.DB = db
	router := setupRouter()
	quizID := 1
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/quiz/%d/1", quizID), nil)
	w:=httptest.NewRecorder()
	router.ServeHTTP(w,req)
	assert.Equal(t, http.StatusOK, w.Code)

	var quizFromDB quizmodels.Quiz
    err = common.DB.First(&quizFromDB, quizID).Error
    if err == nil {
        t.Fatalf("Expected quiz to be deleted, but it was found in the database")
    } else if err != gorm.ErrRecordNotFound {
        t.Fatalf("Failed to check if quiz was deleted: %v", err)
    }
	var department quizmodels.Department
    err = common.DB.Where("name = ?", "Science").First(&department).Error
    if err != nil && err != gorm.ErrRecordNotFound {
        t.Fatalf("Failed to retrieve department: %v", err)
    }
	if err == nil {
		if department.NumQuizzes  !=0 {
			t.Fatalf("Expected department to be deleted, but it still has quizzes")
		}
    }
}
