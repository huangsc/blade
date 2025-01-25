package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huangsc/blade/server/http"
)

func main() {
	// 创建 HTTP 服务器
	server := http.New(
		http.WithAddress("0.0.0.0"),
		http.WithPort(8080),
		http.WithTimeout(time.Second*30),
		http.WithMode(gin.ReleaseMode),
		http.WithHeaders(map[string]string{
			"Server": "Knife/1.0",
		}),
	)

	// 注册路由
	v1 := server.Group("/v1")
	{
		// 健康检查
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "UP",
				"time":   time.Now().Format(time.RFC3339),
			})
		})

		// 用户服务示例
		users := v1.Group("/users")
		{
			users.POST("/", createUser)
			users.GET("/:id", getUser)
			users.PUT("/:id", updateUser)
			users.DELETE("/:id", deleteUser)
		}
	}

	// 启动服务器
	go func() {
		log.Printf("HTTP服务器正在启动，监听地址：http://localhost:8080")
		if err := server.Start(); err != nil {
			log.Printf("HTTP服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	log.Println("正在关闭HTTP服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Printf("HTTP服务器关闭失败: %v", err)
	}
	log.Println("HTTP服务器已关闭")
}

// User 用户模型
type User struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	CreateAt time.Time `json:"create_at"`
	UpdateAt time.Time `json:"update_at"`
}

// 创建用户
func createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 模拟创建用户
	user.ID = "user-123"
	user.CreateAt = time.Now()
	user.UpdateAt = time.Now()

	c.JSON(201, user)
}

// 获取用户
func getUser(c *gin.Context) {
	id := c.Param("id")

	// 模拟查询用户
	user := &User{
		ID:       id,
		Name:     "测试用户",
		Email:    "test@example.com",
		CreateAt: time.Now().Add(-time.Hour * 24),
		UpdateAt: time.Now(),
	}

	c.JSON(200, user)
}

// 更新用户
func updateUser(c *gin.Context) {
	id := c.Param("id")
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user.ID = id
	user.UpdateAt = time.Now()

	c.JSON(200, user)
}

// 删除用户
func deleteUser(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{
		"message": "用户已删除",
		"id":      id,
	})
}
