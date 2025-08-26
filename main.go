// 文件路径: main.go
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gobbs/config"
	"gobbs/logger"
	"gobbs/models" // 导入你刚刚创建的 models 包
	"gobbs/routes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	logger.Init()

	config.LoadConfig()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.AppConfig.MySQL.User,
		config.AppConfig.MySQL.Password,
		config.AppConfig.MySQL.Host,
		config.AppConfig.MySQL.Port,
		config.AppConfig.MySQL.DBName,
	)

	var err error
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		zap.L().Fatal("连接数据库失败", zap.Error(err))
	}
	zap.L().Info("数据库连接成功!")
	db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{})
	zap.L().Info("数据库迁移成功!")

	//2.初始化Gin引擎，注册路由
	r := gin.Default()
	routes.SetupRoutes(r, db)

	//4.启动Web服务
	port := config.AppConfig.Server.Port
	zap.L().Info(fmt.Sprintf("Web 服务启动，监听在%d端口...", port))
	r.Run(fmt.Sprintf(":%d", port))
}
