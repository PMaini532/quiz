package routes

import (
	"net/http"
	"new-mini-project/service2_user/handlers"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context){
		session := sessions.Default(c)
		sessionuserID := session.Get("user_id")
		if sessionuserID == nil{
			c.JSON(http.StatusUnauthorized,gin.H{"error":"Unauthorised"})
			c.Abort()
			return
		}
		urlUserID := c.Param("user_id")
		urluserId,err := strconv.Atoi(urlUserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
        }
		 if sessionuserID != urluserId {
            c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
            c.Abort()
            return
        }
		c.Set("user_id", sessionuserID)
        c.Next()
	}
}

func SetupRoutes(router *gin.Engine)  {

	router.POST("/register",handlers.RegisterUser)
	router.POST("/login",handlers.LoginUser)
	router.GET("/departments/:department",handlers.GetUsersByDepartment)
	authorized := router.Group("/")
	authorized.Use(AuthRequired())
    {
        authorized.PUT("/user/:user_id", handlers.UpdateUser)
        authorized.DELETE("/user/:user_id", handlers.DeleteUser)
    }
	// router.DELETE("/user/:user_id", handlers.DeleteUser)
	router.GET("/users",handlers.GetAllUsers)
	router.GET("/check-session",handlers.CheckSession)
	router.GET("/logout",handlers.Logout)
}
