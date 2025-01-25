package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huangsc/blade/auth"
)

func main() {
	// 创建认证器
	authenticator := auth.NewJWTAuthenticator(
		"your-secret-key",
		time.Hour*24,
	)

	// 创建路由
	r := gin.Default()

	// 登录接口
	r.POST("/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 这里应该验证用户名和密码
		// 示例中简单判断
		if req.Username != "admin" || req.Password != "123456" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// 生成令牌
		claims := auth.Claims{
			UserID:   "1",
			Username: req.Username,
			Role:     "admin",
		}

		token, err := authenticator.GenerateToken(claims)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
		})
	})

	// 受保护的API组
	api := r.Group("/api")
	api.Use(auth.AuthMiddleware(authenticator))
	{
		// 获取用户信息
		api.GET("/user", func(c *gin.Context) {
			claims, _ := c.Get(string(auth.ClaimsKey))
			userClaims := claims.(*auth.Claims)

			c.JSON(http.StatusOK, gin.H{
				"user_id":  userClaims.UserID,
				"username": userClaims.Username,
				"role":     userClaims.Role,
			})
		})

		// 需要管理员权限的接口
		admin := api.Group("/admin")
		admin.Use(auth.RequireRole("admin"))
		{
			admin.POST("/users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "创建用户成功",
				})
			})

			admin.DELETE("/users/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "删除用户成功",
				})
			})
		}
	}

	// 启动服务器
	log.Println("Server is running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
