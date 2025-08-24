package handlers

import (
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"gobbs/models"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func CreatePostHandler(db *gorm.DB, node *snowflake.Node) gin.HandlerFunc {
	return func(c *gin.Context) {
		//从JWT中间件获取当前登录用户的ID
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			return
		}
		userID, ok := userIDValue.(int64)
		if !ok {
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

		communityID, err := strconv.ParseInt(communityIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "板块ID格式错误"})
			return
		}

		//数据入库
		newPost := models.Post{
			PostID:      node.Generate().Int64(),
			AuthorID:    userID,
			CommunityID: communityID,
			Title:       title,
			Content:     content,
		}

		createResult := db.Create(&newPost)
		if createResult.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "帖子创建失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "帖子发布成功", "post_id": newPost.PostID})
	}
}
