package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/huangsc/blade/examples/auth/grpc/proto"
	"github.com/huangsc/blade/auth"
	"google.golang.org/grpc"
)

func main() {
	// 创建认证器
	authenticator := auth.NewJWTAuthenticator(
		"your-secret-key",
		time.Hour*24,
	)

	// 生成测试令牌
	token, err := authenticator.GenerateToken(auth.Claims{
		UserID:   "1",
		Username: "admin",
		Role:     "admin",
	})
	if err != nil {
		log.Fatalf("failed to generate token: %v", err)
	}

	// 创建带认证的连接
	conn, err := grpc.Dial("localhost:9000",
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(auth.UnaryClientInterceptor(token)),
		grpc.WithStreamInterceptor(auth.StreamClientInterceptor(token)),
	)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// 创建客户端
	client := pb.NewUserServiceClient(conn)
	ctx := context.Background()

	// 获取用户信息
	log.Println("\n获取用户信息:")
	user, err := client.GetUser(ctx, &pb.GetUserRequest{Id: "1"})
	if err != nil {
		log.Printf("GetUser failed: %v", err)
	} else {
		log.Printf("User: %+v", user)
	}

	// 获取用户列表
	log.Println("\n获取用户列表:")
	stream, err := client.ListUsers(ctx, &pb.ListUsersRequest{
		PageSize: 10,
		PageNum:  1,
	})
	if err != nil {
		log.Printf("ListUsers failed: %v", err)
	} else {
		for {
			user, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("ListUsers stream failed: %v", err)
				break
			}
			log.Printf("User: %+v", user)
		}
	}

	// 更新用户信息
	log.Println("\n更新用户信息:")
	updatedUser, err := client.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:    "1",
		Email: "new@example.com",
	})
	if err != nil {
		log.Printf("UpdateUser failed: %v", err)
	} else {
		log.Printf("Updated user: %+v", updatedUser)
	}

	// 删除用户
	log.Println("\n删除用户:")
	resp, err := client.DeleteUser(ctx, &pb.DeleteUserRequest{Id: "1"})
	if err != nil {
		log.Printf("DeleteUser failed: %v", err)
	} else {
		log.Printf("Delete response: %+v", resp)
	}
}
