package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/huangsc/blade/examples/tracing/grpc/proto"
	"github.com/huangsc/blade/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	// 创建追踪器
	tracer, err := tracing.NewOTelTracer(
		tracing.WithServiceName("grpc-client"),
		tracing.WithServiceVersion("v1.0.0"),
		tracing.WithEnvironment("development"),
		tracing.WithEndpoint("localhost:4317"),
		tracing.WithSampler(1.0),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// 连接 gRPC 服务器
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewHelloServiceClient(conn)

	// 调用普通 RPC 方法
	callUnaryRPC(tracer, client)

	// 调用流式 RPC 方法
	callStreamingRPC(tracer, client)
}

func callUnaryRPC(tracer tracing.Tracer, client pb.HelloServiceClient) {
	ctx := context.Background()

	// 创建客户端 Span
	ctx, span := tracer.Start(ctx, "grpc.hello.say_hello",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithSpanAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", "HelloService"),
			attribute.String("rpc.method", "SayHello"),
		),
	)
	defer span.End()

	// 注入追踪上下文到 gRPC 元数据
	carrier := propagation.MapCarrier{}
	if err := tracer.Inject(ctx, carrier); err != nil {
		span.SetError(err)
		log.Printf("Failed to inject context: %v", err)
		return
	}

	md := metadata.New(nil)
	for k, v := range carrier {
		md.Set(k, v)
	}
	ctx = metadata.NewOutgoingContext(ctx, md)

	// 发送请求
	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "Alice"})
	if err != nil {
		span.SetError(err)
		log.Printf("Failed to call SayHello: %v", err)
		return
	}

	log.Printf("Response: %s (at %s)", resp.Message, resp.Timestamp)
}

func callStreamingRPC(tracer tracing.Tracer, client pb.HelloServiceClient) {
	ctx := context.Background()

	// 创建客户端 Span
	ctx, span := tracer.Start(ctx, "grpc.hello.say_hello_stream",
		tracing.WithSpanKind(tracing.SpanKindClient),
		tracing.WithSpanAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", "HelloService"),
			attribute.String("rpc.method", "SayHelloStream"),
		),
	)
	defer span.End()

	// 注入追踪上下文到 gRPC 元数据
	carrier := propagation.MapCarrier{}
	if err := tracer.Inject(ctx, carrier); err != nil {
		span.SetError(err)
		log.Printf("Failed to inject context: %v", err)
		return
	}

	md := metadata.New(nil)
	for k, v := range carrier {
		md.Set(k, v)
	}
	ctx = metadata.NewOutgoingContext(ctx, md)

	// 发送请求
	stream, err := client.SayHelloStream(ctx, &pb.HelloRequest{Name: "Bob"})
	if err != nil {
		span.SetError(err)
		log.Printf("Failed to call SayHelloStream: %v", err)
		return
	}

	// 接收响应
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			span.SetError(err)
			log.Printf("Failed to receive response: %v", err)
			return
		}

		log.Printf("Response: %s (at %s)", resp.Message, resp.Timestamp)
		time.Sleep(time.Second)
	}
}
