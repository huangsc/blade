package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/huangsc/blade/database"
	"github.com/huangsc/blade/database/mysql"
)

// User 用户模型
type User struct {
	ID        uint      `gorm:"primarykey"`
	Name      string    `gorm:"size:100;not null"`
	Email     string    `gorm:"size:100;uniqueIndex"`
	Age       int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func main() {
	// 创建数据库实例
	db, err := mysql.NewDB(
		database.WithDSN("root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"),
		database.WithMaxOpenConns(10),
		database.WithMaxIdleConns(5),
		database.WithConnMaxLifetime(time.Hour),
		database.WithSlowThreshold(time.Millisecond*100),
		database.WithTracing(true),
		database.WithLogger(os.Stdout),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// 自动迁移
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal(err)
	}

	// 插入数据
	log.Println("\n测试插入...")
	user := &User{
		Name:      "张三",
		Email:     "zhangsan@example.com",
		Age:       20,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(ctx, user); err != nil {
		log.Fatal(err)
	}
	log.Printf("插入成功，ID=%d\n", user.ID)

	// 查询单条
	log.Println("\n测试查询单条...")
	var foundUser User
	if err := db.First(ctx, &foundUser, user.ID); err != nil {
		log.Fatal(err)
	}
	log.Printf("用户信息: %+v\n", foundUser)

	// 更新数据
	log.Println("\n测试更新...")
	if err := db.Updates(ctx, &foundUser, map[string]interface{}{
		"name":       "李四",
		"age":        21,
		"updated_at": time.Now(),
	}); err != nil {
		log.Fatal(err)
	}
	log.Println("更新成功")

	// 查询多条
	log.Println("\n测试查询多条...")
	var users []User
	if err := db.Find(ctx, &users); err != nil {
		log.Fatal(err)
	}
	log.Printf("所有用户: %+v\n", users)

	// 测试事务
	log.Println("\n测试事务...")
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// 在事务中执行多个操作
	newUser := &User{
		Name:      "王五",
		Email:     "wangwu@example.com",
		Age:       22,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tx.Create(ctx, newUser); err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	if err := tx.Update(ctx, &foundUser, "name", "赵六"); err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	log.Println("事务提交成功")

	// 删除数据
	log.Println("\n测试删除...")
	if err := db.Delete(ctx, &foundUser); err != nil {
		log.Fatal(err)
	}
	log.Println("删除成功")

	// 打印连接池统计信息
	stats := db.Stats()
	log.Printf("\n连接池统计:\n"+
		"最大连接数: %d\n"+
		"打开连接数: %d\n"+
		"使用中连接数: %d\n"+
		"空闲连接数: %d\n"+
		"等待数: %d\n"+
		"等待时间: %v\n"+
		"最大空闲关闭数: %d\n"+
		"最大生命周期关闭数: %d\n",
		stats.MaxOpenConnections,
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		stats.WaitCount,
		stats.WaitDuration,
		stats.MaxIdleClosed,
		stats.MaxLifetimeClosed,
	)
}
