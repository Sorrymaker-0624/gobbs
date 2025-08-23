package middlewares

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gobbs/handlers"
	"net/http"
	"strings"
)

type MyClaims struct {
	UserID   int64  `json:"userID"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		//从请求头获取Token
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请求中缺少Auth Token"})
			c.Abort()
			return
		}

		//解析Token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Auth Token格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("非预期的签名方法")
			}
			return handlers.MySecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
			c.Set("userID", claims.UserID)
			c.Set("username", claims.Username)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token Claims"})
			c.Abort()
			return
		}

		c.Next()
	}

}
