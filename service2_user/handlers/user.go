package handlers

import (
	"log"
	"net/http"
	"new-mini-project/common"
	"new-mini-project/service1_quiz/quizmodels"
	"new-mini-project/service2_user/usermodels"
	"new-mini-project/service3_test/testmodels"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func hashPassword(password string) (string,error) {
	bytes,err := bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	return string(bytes),err
}

func checkPasswordHash(password,hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash),[]byte(password))
	return err == nil
}

func DepartmentExists(departmentname string) bool {
	var department quizmodels.Department
	if err := common.DB.Where("name = ?",departmentname).First(&department).Error; err != nil{
		return false
	}
	return true
}

func getDeptNames() []string {
	var departments []quizmodels.Department
	var deptname []string
	if err := common.DB.Find(&departments).Error; err == nil{
		for _,dep := range departments{
			deptname = append(deptname, dep.Name)
		}
	}
	return deptname
}

func RegisterUser(c *gin.Context)  {
	var user usermodels.User
	err := c.ShouldBindJSON(&user)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
		return
	}
	if user.Email == ""{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Please provide email!"})
		return
	}
	if user.Password == ""{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Please provide password!"})
		return
	}
	isAdmin := user.IsAdmin
	if !isAdmin{
		if user.Department == ""{
			c.JSON(http.StatusBadRequest,gin.H{"error":"Please provide department!"})
			return
		}
		if !DepartmentExists(user.Department)  {
	
			c.JSON(http.StatusBadRequest,gin.H{"error":"Department does not exist.","message":"These are the availabe departments, choose from them:-",
			"departments": getDeptNames()})
			return
		}
	}
	var existingUser usermodels.User
	if err := common.DB.Where("email = ?",user.Email).First(&existingUser).Error; err == nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Email already registered"})
		return
	}
	var deletedID usermodels.DeletedID
	var newID int
	if err := common.DB.Order("id asc").First(&deletedID).Error; err == nil{
		newID = deletedID.ID
		if err := common.DB.Delete(&deletedID).Error; err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove reused ID"})
			return
		}
	}else{
		var lastUser usermodels.User
		if err := common.DB.Last(&lastUser).Error; err != nil {
			newID = 1
		} else {
			newID = lastUser.ID + 1
		}
	} 
	user.ID = newID
	hashedPassword, err := hashPassword(user.Password)
	if err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to hash password"})
		return
	}
	user.Password = hashedPassword
	if err := common.DB.Create(&user).Error; err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to create user"})
		return
	}
	c.JSON(http.StatusOK,user)
}

func LoginUser(c *gin.Context)  {
	var credentials struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}
	err := c.ShouldBindJSON(&credentials)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
		return
	}
	if credentials.Email == ""{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Please provide email!"})
		return
	}
	if credentials.Password == ""{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Please provide password!"})
		return
	}
	var user usermodels.User
	if err := common.DB.Where("email = ?",credentials.Email).First(&user).Error; err != nil{
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}
	if !checkPasswordHash(credentials.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id",user.ID)
	session.Save()
	var department struct {
        Department string `json:"department"`
    }
    if err := common.DB.Model(&user).Select("department").Scan(&department).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch department"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message":    "Login successful",
        "department": department.Department,
    })
}

func GetUsersByDepartment(c *gin.Context)  {
	department := c.Param("department")

	var departmentUsers []usermodels.User
	if err := common.DB.Where("department = ?",department).Find(&departmentUsers).Error; err !=nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to fetch users"})
		return
	}
	if len(departmentUsers) == 0{
		c.JSON(http.StatusNotFound,gin.H{"error":"No users found for this department"})
		return
	}
	c.JSON(http.StatusOK,departmentUsers)
}
func GetAllUsers(c *gin.Context)  {
	var users []usermodels.User
	if err := common.DB.Find(&users).Error; err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}
func UpdateUser(c *gin.Context)  {
	userIDParam := c.Param("user_id")
	userID,err := strconv.Atoi(userIDParam)
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid UserId"})
		return
	}
	var updatedUser usermodels.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil{
		c.JSON(http.StatusBadRequest,gin.H{"error": err.Error()})
		return
	}


	var user usermodels.User
	if err := common.DB.First(&user,userID).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		}else{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}
	if updatedUser.Department != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Department cannot be changed"})
		return
	}
	if updatedUser.IsAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Admin status cannot be changed"})
		return
	}
	
	updateMap := make(map[string]interface{})

	if updatedUser.Password != ""{
		hashedPassword, err := hashPassword(updatedUser.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		updateMap["password"] = hashedPassword
	}

	if updatedUser.Email != "" {
		updateMap["email"] = updatedUser.Email
	}

	if updatedUser.Name != "" {
		updateMap["name"] = updatedUser.Name
	}
	if updatedUser.Department == "" {
		updateMap["department"] = user.Department
	}
	if updatedUser.IsAdmin {
		updateMap["is_admin"] = user.IsAdmin
	}
	if err := common.DB.Model(&user).Updates(updateMap).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	var user usermodels.User
	if err := common.DB.First(&user,userID).Error; err != nil{
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound,gin.H{"error":"User not found"})
		}else{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to retrieve user"})
		}
		return
	}

	tx := common.DB.Begin()
	if tx.Error != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
        return
	}
	if err := tx.Where("user_id = ?",userID).Delete(testmodels.QuizScore{}).Error; err != nil{
		tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete associated scores"})
        return
	}
	if err := tx.Delete(&user).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }

	deletedID := usermodels.DeletedID{ID: userID}
	if err := tx.Create(&deletedID).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record deleted ID"})
        return
    }
	if err := tx.Commit().Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
        return
    }
	session := sessions.Default(c)
    session.Clear() 
    if err := session.Save(); err != nil {
        log.Println("Failed to clear session:", err)
    }
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func CheckSession(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"userID": userID,
	})
}


func Logout(c *gin.Context)  {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil{
		log.Println("Failed to clear session:",err)
	}
	c.JSON(http.StatusOK, gin.H{"message":"Logged Out"})
}