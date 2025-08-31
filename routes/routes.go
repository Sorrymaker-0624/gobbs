package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gobbs/handlers"
	"gobbs/middlewares"
	"gorm.io/gorm"
	"net/http"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client) {
	v1 := r.Group("/api/v1")
	{
		// --- 公开路由 (Public Routes) ---
		// 这一部分接口不需要登录就可以访问
		v1.POST("/register", handlers.RegisterHandler(db))
		v1.POST("/login", handlers.LoginHandler(db, rdb))

		// 查看公开信息
		v1.GET("/users/:username", handlers.GetUserInfoHandler(db))
		v1.GET("/posts", handlers.GetPostListHandler(db))
		v1.GET("/posts/:post_id", handlers.GetPostDetailHandler(db, rdb))
		v1.GET("/posts/:post_id/comments", handlers.GetCommentListHandler(db))

		// 创建一个新的子路由组，并为这个组应用认证中间件
		authed := v1.Group("")
		authed.Use(middlewares.SessionAuthMiddleware(rdb))
		{
			// 在这个花括号里的接口，都必须经过 SessionAuthMiddleware 的验证
			authed.GET("/profile", func(c *gin.Context) {
				userID, _ := c.Get("userID")
				username, _ := c.Get("username")
				c.JSON(http.StatusOK, gin.H{
					"message":  "这是受保护的个人信息接口",
					"userID":   userID,
					"username": username,
				})
			})

			// 创建资源
			authed.POST("/posts", handlers.CreatePostHandler(db))                      // 发布帖子
			authed.POST("/posts/:post_id/comments", handlers.CreateCommentHandler(db)) // 发表评论
			authed.POST("/posts/:post_id/like", handlers.LikePostHandler(db, rdb))     //帖子点赞
			authed.POST("/comments/:comment_id/like", handlers.LikeCommentHandler(db, rdb))
		}
	}
}
