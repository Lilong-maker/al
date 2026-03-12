package main

import (
	pb "al/proto"
	"al/srv/dasic/config"
	"al/srv/dasic/inits"
	"al/srv/dasic/interceptor"
	"al/srv/handler/service"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

// 命令行参数
var (
	port   = flag.Int("port", 50051, "The server port")
	env    = flag.String("env", "", "Environment: dev, beta, prod")
	dataID = flag.String("dataid", "order", "Nacos Data ID")
)

func main() {
	if err := inits.ConsulInit(); err != nil {
		log.Fatalf("Consul初始化失败: %v", err)
	}
	log.Println("Consul初始化成功")
	services, err := inits.GetServiceWithLoadBalancer(config.Gen.ConSul.ServiceName)
	if err != nil {
		log.Printf("获取用户服务失败: %v", err)
	} else {
		log.Printf("获取到用户服务: %s, 地址: %s:%d", services.Service, services.Address, services.Port)
	}

	// 解析命令行参数
	flag.Parse()

	// 设置环境变量（在 inits 执行前设置）
	if *env != "" {
		os.Setenv("APP_ENV", *env)
	}
	if *dataID != "" {
		os.Setenv("APP_DATA_ID", *dataID)
	}

	// 显示当前环境
	currentEnv := os.Getenv("APP_ENV")
	if currentEnv == "" {
		currentEnv = "dev" // 默认使用 dev 环境
	}
	log.Printf("========================================")
	log.Printf("当前运行环境: %s", currentEnv)
	log.Printf("配置 DataID: %s", os.Getenv("APP_DATA_ID"))
	log.Printf("========================================")

	// 启动 gRPC 服务
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 创建 gRPC 服务器（带拦截器）
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(), // 异常恢复
		),
	)

	pb.RegisterOrderServiceServer(s, &service.Server{})
	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	err = inits.ConsulShutdown()
	if err != nil {
		return
	}
	fmt.Println("服务已退出")
}
