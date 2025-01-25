package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/huangsc/blade/examples/tracing/grpc/proto"
	"github.com/huangsc/blade/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type server struct {
	pb.UnimplementedHelloServiceServer
	tracer tracing.Tracer
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	// 从 gRPC 元数据中提取追踪上下文
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	carrier := propagation.MapCarrier{}
	for k, v := range md {
		if len(v) > 0 {
			carrier.Set(k, v[0])
		}
	}
	ctx = s.tracer.Extract(ctx, carrier)

	// 创建服务器端 Span
	ctx, span := s.tracer.Start(ctx, "grpc.hello.say_hello",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithSpanAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", "HelloService"),
			attribute.String("rpc.method", "SayHello"),
			attribute.String("client.name", req.Name),
		),
	)
	defer span.End()

	// 处理请求
	if err := processGRPCRequest(ctx, s.tracer, req.Name); err != nil {
		span.SetError(err)
		return nil, err
	}

	// 返回响应
	return &pb.HelloResponse{
		Message:   fmt.Sprintf("Hello, %s!", req.Name),
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *server) SayHelloStream(req *pb.HelloRequest, stream pb.HelloService_SayHelloStreamServer) error {
	// 从 gRPC 元数据中提取追踪上下文
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		md = metadata.New(nil)
	}
	carrier := propagation.MapCarrier{}
	for k, v := range md {
		if len(v) > 0 {
			carrier.Set(k, v[0])
		}
	}
	ctx := s.tracer.Extract(stream.Context(), carrier)

	// 创建服务器端 Span
	ctx, span := s.tracer.Start(ctx, "grpc.hello.say_hello_stream",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithSpanAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", "HelloService"),
			attribute.String("rpc.method", "SayHelloStream"),
			attribute.String("client.name", req.Name),
		),
	)
	defer span.End()

	// 发送多条响应
	for i := 0; i < 5; i++ {
		// 处理请求
		if err := processGRPCRequest(ctx, s.tracer, req.Name); err != nil {
			span.SetError(err)
			return err
		}

		// 发送响应
		if err := stream.Send(&pb.HelloResponse{
			Message:   fmt.Sprintf("Hello, %s! (%d)", req.Name, i+1),
			Timestamp: time.Now().Format(time.RFC3339),
		}); err != nil {
			span.SetError(err)
			return err
		}

		time.Sleep(time.Second)
	}

	return nil
}

func processGRPCRequest(ctx context.Context, tracer tracing.Tracer, name string) error {
	// 创建处理 Span
	ctx, span := tracer.Start(ctx, "process_grpc_request",
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithSpanAttributes(
			attribute.String("client.name", name),
		),
	)
	defer span.End()

	// 添加处理开始事件
	span.AddEvent("processing.start",
		attribute.String("client.name", name),
	)

	// 模拟处理延迟
	time.Sleep(time.Duration(100) * time.Millisecond)

	// 添加处理结束事件
	span.AddEvent("processing.end",
		attribute.String("status", "success"),
	)

	return nil
}

func main() {
	// 创建追踪器
	tracer, err := tracing.NewOTelTracer(
		tracing.WithServiceName("grpc-server"),
		tracing.WithServiceVersion("v1.0.0"),
		tracing.WithEnvironment("development"),
		tracing.WithEndpoint("localhost:4317"),
		tracing.WithSampler(1.0),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterHelloServiceServer(s, &server{tracer: tracer})

	log.Println("gRPC server is running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
