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
	err = db.AutoMigrate(&quizmodels.Quiz{}, &quizmodels.Department{}, &usermodels.User{},&usermodels.DeletedID{})
	if err != nil {
		return nil, err
	}
	return db, nil
}


func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/register",RegisterUser)
	r.POST("/login",LoginUser)
	r.GET("/departments/:department",GetUsersByDepartment)
	r.PUT("/user/:user_id",UpdateUser)
	r.DELETE("/user/:user_id",DeleteUser)
	return r
}
var userID int
func TestRegisterUser(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	common.DB = db

	router := setupRouter()

	user := usermodels.User{
		Name:       "Test User",
		Email:      "testuser@example.com",
		Password:   "password123",
		IsAdmin: true,
	}
	userJSON, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var registeredUser usermodels.User
	err = json.Unmarshal(w.Body.Bytes(), &registeredUser)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, registeredUser.Email)
}


func TestRegisteNonAdminUser(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	common.DB = db

	router := setupRouter()

	user := usermodels.User{
		Name:       "Test User",
		Email:      "testuser1@example.com",
		Password:   "password123",
		Department: "Science",
		IsAdmin: false,
	}
	userJSON, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var registeredUser usermodels.User
	err = json.Unmarshal(w.Body.Bytes(), &registeredUser)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, registeredUser.Email)
}

func TestUpdateUser(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	common.DB = db
	userID = 4
	router := setupRouter()

	updatedUser := usermodels.User{
		Name:  "Updated User",
		Email: "updateduser@example.com",
	}
	updatedUserJSON, _ := json.Marshal(updatedUser)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/user/%d", userID), bytes.NewBuffer(updatedUserJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resultUser usermodels.User
	err = json.Unmarshal(w.Body.Bytes(), &resultUser)
	assert.NoError(t, err)
	assert.Equal(t, updatedUser.Email, resultUser.Email)
}


func TestDeleteUser(t *testing.T)  {
	db, err := setupTestDB()
	if err != nil{
		t.Fatalf("Failed to setup database %v",err)
	}
	userID = 4
	common.DB = db
	router := setupRouter()
	req, _ := http.NewRequest("DELETE",fmt.Sprintf("/user/%d", userID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "User deleted", result["message"])
}
