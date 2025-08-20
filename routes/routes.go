package routes

import (
	"github.com/gin-gonic/gin"
	"gobbs/handlers"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	userRoutes := r.Group("users")
	{
		userRoutes.POST("/register", handlers.RegisterHandler(db))
		userRoutes.POST("/login", handlers.LoginHandler(db))
		userRoutes.POST("/userpage", handlers.GetUserInfoHandler(db))
	}
}
