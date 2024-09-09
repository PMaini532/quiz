package handlers

import (
	"errors"
	"net/http"
	"new-mini-project/common"
	"new-mini-project/service1_quiz/quizmodels"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"new-mini-project/service2_user/usermodels"
	"new-mini-project/service3_test/testmodels"
)

func IsDepartmentExists(departmentname string) bool {
	for _,dep := range quizmodels.Departments{
		if dep.Name == departmentname{
			return true
		}
	}
	return false
}

func CreateQuiz(c *gin.Context){
	userIdParam := c.Param("user_id")
	userID,err := strconv.Atoi(userIdParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}
	var quiz quizmodels.Quiz
	err = c.ShouldBindJSON(&quiz)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error": err.Error()})
		return
	}
	var user usermodels.User
	if err := common.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create quizzes"})
		return
	}
	if quiz.Department == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Department cannot be empty"})
		return
	}
	var department quizmodels.Department
	if err := common.DB.Where("name = ?", quiz.Department).First(&department).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			newDepartment := quizmodels.Department{Name: quiz.Department,NumQuizzes: 1}
			if err := common.DB.Create(&newDepartment).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create department"})
				return
			}
		}else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check department"})
			return
		}
	}else {
		department.NumQuizzes++
		if err := common.DB.Save(&department).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update department"})
			return
		}
	}
		
	var lastQuiz quizmodels.Quiz
	if err := common.DB.Order("id desc").First(&lastQuiz).Error; err == nil {
		quiz.ID = lastQuiz.ID + 1
	} else {
		quiz.ID = 1
	}
	if err := common.DB.Create(&quiz).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create quiz"})
		return
	}
	c.JSON(http.StatusOK, quiz)
}

func GetAllQuizes(c *gin.Context) {
	var quizzes []quizmodels.Quiz
	if err := common.DB.Preload("Questions.Options").Find(&quizzes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quizzes"})
		return
	}
	c.JSON(http.StatusOK, quizzes)
}


func GetQuiz(c *gin.Context)  {
	id := c.Param("quiz_id")
	quizId,err := strconv.Atoi(id)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid quiz Id"})
		return
	}
	var quiz quizmodels.Quiz
	if err := common.DB.Preload("Questions.Options").First(&quiz,quizId).Error; err != nil{
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
        return
	}
	c.JSON(http.StatusOK, quiz)
}

func UpdateQuiz(c *gin.Context){
	id := c.Param("quiz_id")
	userIDParam := c.Param("user_id")
	quizId, err := strconv.Atoi(id)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid Quiz ID"})
		return
	}
	userID ,err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid userID"})
	}

	var admin usermodels.User
	if err := common.DB.First(&admin,userID).Error; err != nil || !admin.IsAdmin{
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: Only admins can update quizzes"})
        return
	}

	var newQuizData quizmodels.Quiz
	if err := c.ShouldBindJSON(&newQuizData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var quiz quizmodels.Quiz
	if err := common.DB.Preload("Questions.Options").First(&quiz,quizId).Error; err != nil{
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
        return
	}
	if newQuizData.Department != ""{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Cant change department"})
		return
	}

	if newQuizData.Title != ""{
		quiz.Title = newQuizData.Title
	}
	if newQuizData.Description != ""{
		quiz.Description = newQuizData.Description
	}
	if err := common.DB.Save(&quiz).Error; err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quiz"})
        return
	}
	c.JSON(http.StatusOK, quiz)
}

func DeleteQuiz(c *gin.Context)  {
	id := c.Param("quiz_id")
	userIDParam := c.Param("user_id")
	quizId,err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid quizId"})
		return
	}
	userID,err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid userId"})
		return
	}
	var admin usermodels.User
	if err := common.DB.First(&admin,userID).Error; err != nil || !admin.IsAdmin {
		c.JSON(http.StatusForbidden,gin.H{"error":"Access denied: Only admins can delete quizzes"})
		return
	}
	var quiz quizmodels.Quiz
	if err := common.DB.Preload("Questions").First(&quiz,quizId).Error; err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Quiz not found"})
		return
	}
	for _, question := range quiz.Questions {
        if err := common.DB.Where("question_id = ?", question.ID).Delete(&quizmodels.Option{}).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete options", "details": err.Error()})
            return
        }
    }

	if err := common.DB.Where("quiz_id = ?", quizId).Delete(&quizmodels.Question{}).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete questions", "details": err.Error()})
        return
    }
	if err := common.DB.Where("quiz_id = ?", quizId).Delete(&testmodels.QuizScore{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete associated quiz scores"})
		return
	}
	var department quizmodels.Department
	if err := common.DB.Where("name = ?", quiz.Department).First(&department).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find department"})
		return
	}

	if err := common.DB.Delete(&quiz).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to delete quiz", "details": err.Error()})
		return
	}

	department.NumQuizzes--
	if department.NumQuizzes == 0 {
		if err := common.DB.Delete(&department).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete department"})
			return
		}
	} else {
		if err := common.DB.Save(&department).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update department"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message":"Quiz deleted"})
}

func GetQuestionsByQuiz(c *gin.Context){
	quizIDParam := c.Param("quiz_id")
	quizID,err := strconv.Atoi(quizIDParam)
	if err != nil{
		c.JSON(http.StatusNotFound,gin.H{"error":"Invalid Quiz ID"})
		return
	}
	var quiz quizmodels.Quiz
	 if err := common.DB.Preload("Questions.Options").First(&quiz, quizID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quiz", "details": err.Error()})
        }
        return
	}
	if len(quiz.Questions) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "No questions found for this quiz"})
        return
    }

    c.JSON(http.StatusOK, quiz.Questions)
}

func GetDepartments(c *gin.Context) {
	var departments []quizmodels.Department
	if err := common.DB.Find(&departments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve departments"})
		return
	}
	c.JSON(http.StatusOK, departments)
}



func AddQuestion(c *gin.Context)  {
	quizIDParam := c.Param("quiz_id")
	userIDParam := c.Param("user_id")
	quizID,err := strconv.Atoi(quizIDParam)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid quizId"})
		return
	}
	userID,err := strconv.Atoi(userIDParam)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid user id"})
		return
	}
	var admin usermodels.User
	if err := common.DB.First(&admin, userID).Error; err != nil || !admin.IsAdmin {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can add questions"})
        return
    }

	var question quizmodels.Question
    if err := c.ShouldBindJSON(&question); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if len(question.Options) != 4 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Exactly 4 options are required"})
        return
    }
	var quiz quizmodels.Quiz
    if err := common.DB.First(&quiz, quizID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
        return
    }

    question.QuizID = quizID
    if err := common.DB.Create(&question).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question", "details": err.Error()})
        return
    }
	var options []quizmodels.Option
	for i := range question.Options {
        question.Options[i].QuestionID = question.ID 
		options = append(options, question.Options[i])
    }
	question.Options = options
    c.JSON(http.StatusOK, question)
}

func GetQuestion(c *gin.Context)  {
	quizIDParam := c.Param("quiz_id")
	questionIDParam := c.Param("question_id")

	quizID,err := strconv.Atoi(quizIDParam)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid Quiz Id"})
		return
	}
	questionID, err1 := strconv.Atoi(questionIDParam)
	if err1 != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid Question Id"})
		return
	}
	var question quizmodels.Question
	if err := common.DB.Preload("Options").Where("id = ? AND quiz_id = ?",questionID,quizID).First(&question).Error; err != nil{
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve question", "details": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, question)
}

func UpdateQuestion(c *gin.Context)  {
	userIDParam := c.Param("user_id")
	userID,err := strconv.Atoi(userIDParam)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid UserId"})
		return
	}
	quizIDParam := c.Param("quiz_id")
	questionIDParam := c.Param("question_id")

	quizId,err := strconv.Atoi(quizIDParam)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid Quiz ID"})
	}
	questionID ,err1 := strconv.Atoi(questionIDParam)
	if err1 != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error" : "Invalid Question ID"})
		return
	}
	var admin usermodels.User
    if err := common.DB.First(&admin, userID).Error; err != nil || !admin.IsAdmin {
        c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update questions"})
        return
    }
	var question quizmodels.Question
    if err := common.DB.Preload("Options").Where("id = ? AND quiz_id = ?", questionID, quizId).First(&question).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve question", "details": err.Error()})
        }
        return
    }

	var updatedQuestionData quizmodels.Question
	if err := c.ShouldBindJSON(&updatedQuestionData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updatedQuestionData.Text != ""{
		question.Text = updatedQuestionData.Text 
	}
	if updatedQuestionData.Answer != ""{
		question.Answer = updatedQuestionData.Answer
	}
	if len(updatedQuestionData.Options) > 0{
		optionMap := make(map[int]quizmodels.Option)
        for _, opt := range updatedQuestionData.Options {
            optionMap[opt.ID] = opt
        }
		for i, opt := range question.Options{
			if newOpt,exists := optionMap[opt.ID]; exists{
				if newOpt.Text != ""{
					opt.Text = newOpt.Text
				}
				if newOpt.IsCorrect { 
                    opt.IsCorrect = newOpt.IsCorrect
                }
				opt.Text = newOpt.Text
                question.Options[i] = opt
                delete(optionMap, opt.ID)
			}
		}
	}
	if err := common.DB.Transaction(func (tx *gorm.DB) error {
		if err := tx.Save(&question).Error; err != nil {
            return err
        }
		for _, opt := range question.Options {
            if err := tx.Save(&opt).Error; err != nil {
                return err
            }
        }

        return nil
	});err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question", "details": err.Error()})
        return
    }
	c.JSON(http.StatusOK, question)
}

func DeleteQuestion(c *gin.Context)  {
	quizIDParam := c.Param("quiz_id")
	questionIDParam := c.Param("question_id")

	quizID, err := strconv.Atoi(quizIDParam)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid Quiz ID"})
	}
	questionID, err1 := strconv.Atoi(questionIDParam)
	if err1 != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid Question ID"})
	}

	var question quizmodels.Question
	if err := common.DB.Preload("Options").Where("id = ? AND quiz_id = ?",questionID,quizID).First(&question).Error; err != nil{
		if errors.Is(err,gorm.ErrRecordNotFound){
			c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		}else{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve question", "details": err.Error()})
		}
		return
	}
	if err := common.DB.Transaction(func (tx *gorm.DB) error {
		if err := tx.Delete(&question.Options).Error; err != nil {
            return err
        }
		if err := tx.Delete(&question).Error; err != nil {
            return err
        }
		return nil
	});err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete question", "details": err.Error()})
        return
    }
	c.JSON(http.StatusOK, gin.H{"message": "Question deleted"})
}

// func GetQuizzesbyDepartment(c *gin.Context){
// 	dapartmentName := c.Param("department")
// 	var quizzes []quizmodels.Quiz
// }

func GetQuizzesByDepartmentName(c *gin.Context) {
	departmentName := c.Param("department_name")
	var quizzes []quizmodels.Quiz

	// Fetch quizzes based on the department name
	err := quizmodels.GetQuizzesByDepartmentName(departmentName, &quizzes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quizzes"})
		return
	}

	if len(quizzes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No quizzes found for the specified department"})
		return
	}

	c.JSON(http.StatusOK, quizzes)
}