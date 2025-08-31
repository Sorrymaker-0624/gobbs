package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gobbs/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"time"
)

var MySecret = []byte("这是一个安全的密钥")

type MyClaims struct {
	ID       uint   `json:"userID"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// 注册接口
func RegisterHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		confirmPassword := c.PostForm("confirm_password")
		phone := c.PostForm("phone")
		email := c.PostForm("email")

		if len(username) == 0 || len(password) == 0 || len(email) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户名、密码和邮箱不能为空"})
			return
		}

		if password != confirmPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "两次输入的密码不一致"})
			return
		}

		var user models.User
		resultByUser := db.Where("username = ?", username).First(&user)
		if !errors.Is(resultByUser.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
			return
		}

		var userByEmail models.User
		resultByEmail := db.Where("email = ?", email).First(&userByEmail)
		if !errors.Is(resultByEmail.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusConflict, gin.H{"error": "该邮箱已被注册"})
			return
		}

		if len(phone) > 0 {
			var userByPhone models.User
			resultByPhone := db.Where("phone = ?", phone).First(&userByPhone)
			if !errors.Is(resultByPhone.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusConflict, gin.H{"error": "该手机号已被注册"})
				return
			}
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		newUser := models.User{
			Username: username,
			Password: string(hashedPassword),
			Phone:    phone,
			Email:    email,
		}
		createResult := db.Create(&newUser)
		if createResult.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "用户创建失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
	}
}

// 登录接口
func LoginHandler(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		if len(username) == 0 || len(password) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户名或密码不正确"})
			return
		}
		//查询用户
		var user models.User
		result := db.Where("username = ? OR email = ?", username, username).First(&user)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或密码错误"})
			return
		}

		//校验密码
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或密码错误"})
			return
		}
		//密码正确
		sessionID := uuid.New().String()

		sessionData := map[string]interface{}{
			"userID":   user.ID,
			"username": user.Username,
		}
		sessionDataBytes, _ := json.Marshal(sessionData)

		err = rdb.Set(context.Background(), "session:"+sessionID, sessionDataBytes, 24*time.Hour).Err()
		if err != nil {
			zap.L().Error("创建Session失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
			return
		}
		//登录成功
		c.JSON(http.StatusOK, gin.H{
			"message":    "登录成功",
			"session_id": sessionID,
		})
	}
}

// 获取用户信息
func GetUserInfoHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")

		var user models.User
		result := db.Where("username = ?", username).First(&user)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"username": user.Username,
			"email":    user.Email,
		})
	}
}
