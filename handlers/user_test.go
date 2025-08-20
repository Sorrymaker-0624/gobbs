// 文件路径: handlers/user_test.go
package handlers

import (
	"gobbs/models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupTestDBAndRouter 函数保持不变，它为我们提供了干净的测试环境
func setupTestDBAndRouter() (*gorm.DB, *gin.Engine) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		panic("无法连接到测试数据库: " + err.Error())
	}
	db.AutoMigrate(&models.User{})
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/register", RegisterHandler(db))
	return db, router
}

// TestRegisterHandler 更新了测试用例以匹配最新的注册逻辑
func TestRegisterHandler(t *testing.T) {
	t.Run("成功注册 - 包含所有字段", func(t *testing.T) {
		// 1. 准备
		db, router := setupTestDBAndRouter()
		formData := url.Values{}
		// [修改] 添加了 email 字段
		formData.Set("username", "test_success")
		formData.Set("password", "password123")
		formData.Set("confirm_password", "password123")
		formData.Set("email", "success@example.com")
		formData.Set("phone", "1234567890") // 包含可选的手机号

		// 2. 执行
		req, _ := http.NewRequest("POST", "/register", strings.NewReader(formData.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 3. 断言
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "注册成功")

		// 检查数据库
		var user models.User
		db.First(&user, "username = ?", "test_success")
		assert.Equal(t, "test_success", user.Username)
		// [修改] 检查 email 和 phone 是否也正确保存
		assert.Equal(t, "success@example.com", user.Email)
		assert.Equal(t, "1234567890", user.Phone)
	})

	t.Run("成功注册 - 不包含可选的手机号", func(t *testing.T) {
		// 1. 准备
		_, router := setupTestDBAndRouter()
		formData := url.Values{}
		formData.Set("username", "test_no_phone")
		formData.Set("password", "password123")
		formData.Set("confirm_password", "password123")
		formData.Set("email", "no_phone@example.com")
		// phone 字段不提供

		// 2. 执行 & 3. 断言
		req, _ := http.NewRequest("POST", "/register", strings.NewReader(formData.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "注册成功")
	})

	t.Run("失败 - 用户名已存在", func(t *testing.T) {
		// 1. 准备
		db, router := setupTestDBAndRouter()
		db.Create(&models.User{Username: "existing_user", Email: "a@a.com"}) // 预埋用户

		formData := url.Values{}
		formData.Set("username", "existing_user") // 使用已存在的用户名
		formData.Set("password", "password123")
		formData.Set("confirm_password", "password123")
		formData.Set("email", "new_email@example.com") // 使用新的邮箱

		// 2. 执行 & 3. 断言
		req, _ := http.NewRequest("POST", "/register", strings.NewReader(formData.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "用户名已存在")
	})

	t.Run("失败 - 邮箱已存在", func(t *testing.T) { // [新增] 测试用例
		// 1. 准备
		db, router := setupTestDBAndRouter()
		db.Create(&models.User{Username: "user1", Email: "existing_email@example.com"})

		formData := url.Values{}
		formData.Set("username", "new_user") // 使用新的用户名
		formData.Set("password", "password123")
		formData.Set("confirm_password", "password123")
		formData.Set("email", "existing_email@example.com") // 使用已存在的邮箱

		// 2. 执行 & 3. 断言
		req, _ := http.NewRequest("POST", "/register", strings.NewReader(formData.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "该邮箱已被注册")
	})

	t.Run("失败 - 密码不匹配", func(t *testing.T) { // [新增] 测试用例
		// 1. 准备
		_, router := setupTestDBAndRouter()
		formData := url.Values{}
		formData.Set("username", "test_pwd_mismatch")
		formData.Set("password", "password123")
		formData.Set("confirm_password", "password456") // 两次密码不一致
		formData.Set("email", "test@example.com")

		// 2. 执行 & 3. 断言
		req, _ := http.NewRequest("POST", "/register", strings.NewReader(formData.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "两次输入的密码不一致")
	})

	t.Run("失败 - 缺少必填项(邮箱)", func(t *testing.T) {
		// 1. 准备
		_, router := setupTestDBAndRouter()
		formData := url.Values{}
		formData.Set("username", "test_missing_field")
		formData.Set("password", "password123")
		formData.Set("confirm_password", "password123")
		// email 字段不提供

		// 2. 执行 & 3. 断言
		req, _ := http.NewRequest("POST", "/register", strings.NewReader(formData.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "用户名、密码和邮箱不能为空")
	})
}
