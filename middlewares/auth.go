package middlewares

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strings"
)

//	func JWTAuthMiddleware() func(c *gin.Context) {
//		return func(c *gin.Context) {
//			//从请求头获取Token
//			authHeader := c.Request.Header.Get("Authorization")
//			if authHeader == "" {
//				c.JSON(http.StatusUnauthorized, gin.H{"error": "请求中缺少Auth Token"})
//				c.Abort()
//				return
//			}
//
//			//解析Token格式
//			parts := strings.SplitN(authHeader, " ", 2)
//			if !(len(parts) == 2 && parts[0] == "Bearer") {
//				c.JSON(http.StatusUnauthorized, gin.H{"error": "Auth Token格式错误"})
//				c.Abort()
//				return
//			}
//
//			tokenString := parts[1]
//			token, err := jwt.ParseWithClaims(tokenString, &handlers.MyClaims{}, func(token *jwt.Token) (interface{}, error) {
//				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
//					return nil, errors.New("非预期的签名方法")
//				}
//				return handlers.MySecret, nil
//			})
//
//			if err != nil {
//				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token"})
//				c.Abort()
//				return
//			}
//
//			if claims, ok := token.Claims.(*handlers.MyClaims); ok && token.Valid {
//				c.Set("userID", claims.ID)
//				c.Set("username", claims.Username)
//			} else {
//				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token Claims"})
//				c.Abort()
//				return
//			}
//
//			c.Next()
//		}
//	}
func SessionAuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证信息"})
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证信息格式错误"})
			c.Abort()
			return
		}
		sessionID := parts[1]

		sessionDataBytes, err := rdb.Get(context.Background(), "session :"+sessionID).Bytes()
		if err == redis.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Session或已过期"})
			c.Abort()
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Session查询失败"})
			c.Abort()
			return
		}

		var sessionData map[string]interface{}
		json.Unmarshal(sessionDataBytes, &sessionData)

		c.Set("userID", uint(sessionData["userID"].(float64)))
		c.Set("username", sessionData["username"].(string))

		c.Next()
	}
}
