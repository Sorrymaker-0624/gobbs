package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gobbs/models"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

func CreatePostHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		//从JWT中间件获取当前登录用户的ID
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			return
		}
		userID, ok := userIDValue.(uint)
		if !ok {
			zap.L().Error("无法转换用户ID", zap.Any("userIDValue", userIDValue))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法转换用户ID"})
			return
		}
		//获取前端传来的参数
		title := c.PostForm("title")
		content := c.PostForm("content")
		communityIDStr := c.PostForm("community_id")
		//参数校验
		if len(title) == 0 || len(content) == 0 || len(communityIDStr) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "标题、内容或板块ID不能为空"})
			return
		}

		communityID, err := strconv.ParseUint(communityIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "板块ID格式错误"})
			return
		}

		//数据入库
		newPost := models.Post{
			AuthorID:    userID,
			CommunityID: uint(communityID),
			Title:       title,
			Content:     content,
		}

		createResult := db.Create(&newPost)
		if createResult.Error != nil {
			zap.L().Error("帖子创建失败", zap.Error(createResult.Error))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "帖子创建失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "帖子发布成功", "post_id": newPost.ID})
	}
}

func GetPostListHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		//获取分页参数
		pageStr := c.DefaultQuery("page", "1")
		sizeStr := c.DefaultQuery("size", "10")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		size, err := strconv.Atoi(sizeStr)
		if err != nil || size < 1 {
			size = 10
		}

		offset := (page - 1) * size

		var posts []models.Post
		result := db.Order("created_at DESC").Offset(offset).Limit(size).Find(&posts)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询帖子列表失败"})
			return
		}

		c.JSON(http.StatusOK, posts)
	}
}

type PostDetailResponse struct {
	ID          uint      `json:"id"`
	AuthorID    uint      `json:"author_id"`
	CommunityID uint      `json:"community_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	AuthorName  string    `json:"author_name"` // 附带上作者名
}

func GetPostDetailHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		postIDStr := c.Param("post_id")

		if len(postIDStr) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID不能为空"})
			return
		}
		postID, err := strconv.ParseUint(postIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "帖子ID格式错误"})
			return
		}

		var post models.Post
		result := db.Where("id = ?", postID).Preload("User").First(&post)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
			return
		}

		response := PostDetailResponse{
			ID:          post.ID,
			AuthorID:    post.AuthorID,
			CommunityID: post.CommunityID,
			Title:       post.Title,
			Content:     post.Content,
			CreatedAt:   post.CreatedAt,
			AuthorName:  post.User.Username, // 从预加载的User对象中获取用户名
		}
		c.JSON(http.StatusOK, response)
	}
}
