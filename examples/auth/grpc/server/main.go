package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "github.com/huangsc/blade/examples/auth/grpc/proto"
	"github.com/huangsc/blade/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// userService 实现用户服务
type userService struct {
	pb.UnimplementedUserServiceServer
}

// GetUser 获取用户信息
func (s *userService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	// 从上下文获取认证信息
	claims, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get claims from context")
	}

	// 检查是否是请求自己的信息
	if req.Id != claims.UserID && claims.Role != "admin" {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	// 模拟从数据库获取用户
	return &pb.User{
		Id:        req.Id,
		Username:  "test_user",
		Email:     "test@example.com",
		Role:      "user",
		CreatedAt: time.Now().Add(-24 * time.Hour).Unix(),
		UpdatedAt: time.Now().Unix(),
	}, nil
}

// ListUsers 获取用户列表
func (s *userService) ListUsers(req *pb.ListUsersRequest, stream pb.UserService_ListUsersServer) error {
	// 从上下文获取认证信息
	claims, ok := auth.FromContext(stream.Context())
	if !ok {
		return status.Error(codes.Internal, "failed to get claims from context")
	}

	// 只允许管理员访问
	if claims.Role != "admin" {
		return status.Error(codes.PermissionDenied, "permission denied")
	}

	// 模拟流式返回用户列表
	for i := 0; i < 5; i++ {
		user := &pb.User{
			Id:        "user-" + string(rune(i+'1')),
			Username:  "user_" + string(rune(i+'1')),
			Email:     "user" + string(rune(i+'1')) + "@example.com",
			Role:      "user",
			CreatedAt: time.Now().Add(-24 * time.Hour).Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		if err := stream.Send(user); err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 100) // 模拟延迟
	}
	return nil
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	// 从上下文获取认证信息
	claims, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get claims from context")
	}

	// 检查是否是更新自己的信息
	if req.Id != claims.UserID && claims.Role != "admin" {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	// 模拟更新用户
	return &pb.User{
		Id:        req.Id,
		Username:  "test_user",
		Email:     req.Email,
		Role:      "user",
		CreatedAt: time.Now().Add(-24 * time.Hour).Unix(),
		UpdatedAt: time.Now().Unix(),
	}, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	// 从上下文获取认证信息
	claims, ok := auth.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get claims from context")
	}

	// 只允许管理员删除用户
	if claims.Role != "admin" {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	// 模拟删除用户
	return &pb.DeleteUserResponse{
		Message: "User deleted successfully",
	}, nil
}

func main() {
	// 创建认证器
	authenticator := auth.NewJWTAuthenticator(
		"your-secret-key",
		time.Hour*24,
	)

	// 创建 gRPC 服务器
	server := grpc.NewServer(
		grpc.UnaryInterceptor(auth.UnaryServerInterceptor(authenticator)),
		grpc.StreamInterceptor(auth.StreamServerInterceptor(authenticator)),
	)

	// 注册服务
	pb.RegisterUserServiceServer(server, &userService{})

	// 启动服务器
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("gRPC server is running on :9000")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
