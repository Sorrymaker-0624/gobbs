package routes

import (
	"github.com/gin-gonic/gin"
	"gobbs/handlers"
	"gobbs/middlewares"
	"gorm.io/gorm"
	"net/http"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	r.POST("/register", handlers.RegisterHandler(db))
	r.POST("/login", handlers.LoginHandler(db))
	r.GET("/users/:username", handlers.GetUserInfoHandler(db))
	r.GET("/posts", handlers.GetPostListHandler(db))
	r.GET("/posts/:post_id", handlers.GetPostDetailHandler(db))

	v1 := r.Group("/api/v1")
	v1.Use(middlewares.JWTAuthMiddleware())
	{
		v1.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			username, _ := c.Get("username")
			c.JSON(http.StatusOK, gin.H{
				"message":  "这是受保护的个人信息接口",
				"userID":   userID,
				"username": username,
			})
		})
		v1.POST("/posts", handlers.CreatePostHandler(db))
	}
}
