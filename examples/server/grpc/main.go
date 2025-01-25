package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/huangsc/blade/examples/server/grpc/proto"
	"github.com/huangsc/blade/server/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// userService 实现用户服务
type userService struct {
	pb.UnimplementedUserServiceServer
}

// CreateUser 创建用户
func (s *userService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	// 模拟创建用户
	now := time.Now().Unix()
	return &pb.User{
		Id:       "user-123",
		Name:     req.Name,
		Email:    req.Email,
		CreateAt: now,
		UpdateAt: now,
	}, nil
}

// GetUser 获取用户
func (s *userService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	// 模拟查询用户
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}

	now := time.Now().Unix()
	return &pb.User{
		Id:       req.Id,
		Name:     "测试用户",
		Email:    "test@example.com",
		CreateAt: now - 86400, // 一天前
		UpdateAt: now,
	}, nil
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}

	// 模拟更新用户
	return &pb.User{
		Id:       req.Id,
		Name:     req.Name,
		Email:    req.Email,
		CreateAt: time.Now().Unix() - 86400, // 一天前
		UpdateAt: time.Now().Unix(),
	}, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}

	// 模拟删除用户
	return &pb.DeleteUserResponse{
		Message: "用户已删除",
		Id:      req.Id,
	}, nil
}

func main() {
	// 创建 gRPC 服务器
	server := grpc.New(
		grpc.WithAddress("0.0.0.0"),
		grpc.WithPort(9000),
		grpc.WithTimeout(time.Second*30),
		grpc.WithHealth(true),
		grpc.WithReflection(true),
	)

	// 注册用户服务
	pb.RegisterUserServiceServer(server.Server, &userService{})

	// 启动服务器
	go func() {
		log.Printf("gRPC服务器正在启动，监听地址：localhost:9000")
		if err := server.Start(); err != nil {
			log.Printf("gRPC服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	log.Println("正在关闭gRPC服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Printf("gRPC服务器关闭失败: %v", err)
	}
	log.Println("gRPC服务器已关闭")
}
