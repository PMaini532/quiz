package main

import (
	"log"
	"new-mini-project/common"
	"new-mini-project/service2_user/routes"
	"new-mini-project/service2_user/usermodels"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main()  {

	dsn := "host=db user=maini password=pratham dbname=quizdb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// // Assign the database instance to the models package
	common.DB = db

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database connection object:", err)
	}
	err = sqlDB.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Successfully connected to the database")

	// // Migrate models
	err = db.AutoMigrate( &usermodels.User{},&usermodels.DeletedID{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Replace with your frontend origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("my-session",store))
	routes.SetupRoutes(router)
	router.Run(":8023")
}

