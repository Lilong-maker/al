package main

import (
	_ "al/srv/dasic/inits"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/goherui/gospacex/web"
)

//// 命令行参数
//var (
//	port   = flag.Int("port", 50052, "The server port")
//	env    = flag.String("env", "", "Environment: dev, beta, prod")
//	dataID = flag.String("dataid", "order", "Nacos Data ID")
//)
//
//func main() {
//	if err := inits.ConsulInit(); err != nil {
//		log.Fatalf("Consul初始化失败: %v", err)
//	}
//	log.Println("Consul初始化成功")
//	services, err := inits.GetServiceWithLoadBalancer(config.Gen.ConSul.ServiceName)
//	if err != nil {
//		log.Printf("获取用户服务失败: %v", err)
//	} else {
//		log.Printf("获取到用户服务: %s, 地址: %s:%d", services.Service, services.Address, services.Port)
//	}
//
//	// 解析命令行参数
//	flag.Parse()
//
//	// 设置环境变量（在 inits 执行前设置）
//	if *env != "" {
//		os.Setenv("APP_ENV", *env)
//	}
//	if *dataID != "" {
//		os.Setenv("APP_DATA_ID", *dataID)
//	}
//
//	// 显示当前环境
//	currentEnv := os.Getenv("APP_ENV")
//	if currentEnv == "" {
//		currentEnv = "dev" // 默认使用 dev 环境
//	}
//	log.Printf("========================================")
//	log.Printf("当前运行环境: %s", currentEnv)
//	log.Printf("配置 DataID: %s", os.Getenv("APP_DATA_ID"))
//	log.Printf("========================================")
//
//	// 启动 gRPC 服务
//	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
//	if err != nil {
//		log.Fatalf("failed to listen: %v", err)
//	}
//
//	// 创建 gRPC 服务器（带拦截器）
//	s := grpc.NewServer(
//		grpc.ChainUnaryInterceptor(
//			interceptor.RecoveryInterceptor(), // 异常恢复
//		),
//	)
//
//	__.RegisterOrderServiceServer(s, &service.Server{})
//	log.Printf("server listening at %v", lis.Addr())
//
//	if err := s.Serve(lis); err != nil {
//		log.Fatalf("failed to serve: %v", err)
//	}
//}

func main() {
	// 创建web服务工厂
	factory := web.NewWebServiceFactory()
	// 创建并启动所有四种服务
	services := []web.WebService{}
	// HTTP服务 (8080端口)
	httpService, err := factory.CreateService("http", 8080)
	if err != nil {
		log.Fatalf("创建HTTP服务失败: %v", err)
	}
	services = append(services, httpService)
	if err := httpService.Start(); err != nil {
		log.Fatalf("启动HTTP服务失败: %v", err)
	}
	// HTTPS服务 (8443端口)
	httpsService, err := factory.CreateService("https", 8443)
	if err != nil {
		log.Fatalf("创建HTTPS服务失败: %v", err)
	}
	services = append(services, httpsService)
	if err := httpsService.Start(); err != nil {
		log.Fatalf("启动HTTPS服务失败: %v", err)
	}
	// gRPC服务 (50051端口)
	grpcService, err := factory.CreateService("grpc", 50052)
	if err != nil {
		log.Fatalf("创建gRPC服务失败: %v", err)
	}
	services = append(services, grpcService)
	if err := grpcService.Start(); err != nil {
		log.Fatalf("启动gRPC服务失败: %v", err)
	}
	// WebSocket服务 (8081端口)
	websocketService, err := factory.CreateService("websocket", 8081)
	if err != nil {
		log.Fatalf("创建WebSocket服务失败: %v", err)
	}
	services = append(services, websocketService)
	if err := websocketService.Start(); err != nil {
		log.Fatalf("启动WebSocket服务失败: %v", err)
	}
	log.Println("所有服务已启动")
	// 等待系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	// 停止所有服务
	for _, service := range services {
		if err := service.Stop(); err != nil {
			log.Printf("停止服务失败: %v", err)
		}
	}
	log.Println("所有服务已停止")
}
