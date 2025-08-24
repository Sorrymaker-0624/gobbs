// 文件路径: main.go
package main

import (
	"fmt"
	"github.com/bwmarrin/snowflake"
	"gobbs/models" // 导入你刚刚创建的 models 包
	"gobbs/routes"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:tkc04624@tcp(127.0.0.1:3306)/gobbs?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败, err: %v\n", err)
	}
	fmt.Println("数据库连接成功!")
	db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{})
	fmt.Println("数据库迁移成功!")

	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatalf("雪花算法节点创建失败, err: %v\n", err)
	}
	fmt.Println("雪花算法节点创建成功!")
	//2.初始化Gin引擎
	r := gin.Default()

	//3.注册路由（调用routes包里的函数，并把db对象传进去）
	routes.SetupRoutes(r, db, node)

	//4.启动Web服务
	fmt.Println("Web 服务启动，监听在8080端口...")
	r.Run(":8080")

	// --- CRUD 操作练习部分 ---
	//fmt.Println("--- 开始CRUD操作练习 ---")

	//// 1. 【增 Create】
	//// 创建一个用户实例
	//u1 := models.User{UserID: 101, Username: "gorm_practice_user", Password: "abc"}
	//result := db.Create(&u1)
	//if result.Error != nil {
	//	fmt.Printf("创建用户失败, err: %v\n", result.Error)
	//	return
	//}
	//fmt.Printf("✅ (Create) 创建用户成功, 用户名: %s, ID: %d\n", u1.Username, u1.ID)
	//
	//// 2. 【查 Read】
	//var user models.User
	//db.First(&user, "username = ?", "gorm_practice_user")
	//fmt.Printf("✅ (Read) 查询到用户: %s, 密码: %s\n", user.Username, user.Password)
	//
	//// 3. 【改 Update】
	//db.Model(&user).Update("Password", "xyz")
	//fmt.Printf("✅ (Update) 更新用户 %s 的密码为: %s\n", user.Username, "xyz")
	//
	//// 4. 【删 Delete】
	//db.Delete(&user)
	//fmt.Printf("✅ (Delete) 删除用户: %s\n", user.Username)
	//
	//fmt.Println("--- GORM CRUD 操作练习结束 ---")
}
