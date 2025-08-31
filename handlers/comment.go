package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gobbs/models"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

func CreateCommentHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			return
		}
		userID, ok := userIDValue.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法转换用户ID"})
			return
		}

		postIDStr := c.Param("post_id")
		content := c.PostForm("content")

		postID, err := strconv.ParseUint(postIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID格式错误"})
			return
		}
		if len(content) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容不能为空"})
			return
		}
		newComment := models.Comment{
			PostID:   uint(postID),
			AuthorID: userID,
			Content:  content,
		}

		createResult := db.Create(&newComment)
		if createResult.Error != nil {
			zap.L().Error("评论创建失败", zap.Error(createResult.Error))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "评论创建失败，可能帖子不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "评论发表成功"})
	}
}

type CommentResponse struct {
	ID         uint      `json:"id"`
	PostID     uint      `json:"post_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	AuthorName string    `json:"author_name"`
}

func GetCommentListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		postIDStr := c.Param("post_id")
		postID, err := strconv.ParseUint(postIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID格式错误"})
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
		if page < 1 {
			page = 1
		}
		if size < 1 {
			size = 10
		}
		offset := (page - 1) * size

		var comments []models.Comment
		result := db.Where("post_id = ?", postID).
			Order("created_at ASC").
			Offset(offset).
			Limit(size).
			Preload("User").
			Find(&comments)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询评论列表失败"})
			return
		}

		var response []CommentResponse
		for _, comment := range comments {
			response = append(response, CommentResponse{
				ID:         comment.ID,
				PostID:     comment.PostID,
				Content:    comment.Content,
				CreatedAt:  comment.CreatedAt,
				AuthorName: comment.User.Username, // 从预加载的User对象中获取用户名
			})
		}

		c.JSON(http.StatusOK, response)
	}
}

func LikeCommentHandler(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentIDStr := c.Param("comment_id")
		commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "评论ID格式错误"})
			return
		}

		userIDValue, _ := c.Get("userID")
		userID := userIDValue.(uint)

		redisKey := fmt.Sprintf("comment:likes:%d", commentID)
		isMember, err := rdb.SIsMember(context.Background(), redisKey, userID).Result()
		if err != nil {
			zap.L().Error("Redis查询评论点赞状态失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
			return
		}

		var message string
		var liked bool

		if isMember {
			rdb.SRem(context.Background(), redisKey, userID)
			message = "取消点赞成功"
			liked = false
		} else {
			rdb.SAdd(context.Background(), redisKey, userID)
			message = "点赞成功"
			liked = true
		}

		likesCount, _ := rdb.SCard(context.Background(), redisKey).Result()

		c.JSON(http.StatusOK, gin.H{
			"message": message,
			"likes":   likesCount,
			"liked":   liked,
		})
	}
}
