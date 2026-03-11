package main

import (
	__ "al/proto"
	_ "al/srv/dasic/init"
	"al/srv/dasic/interceptor"
	"al/srv/dasic/metrics"
	"al/srv/handler/service"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

// 命令行参数
var (
	port     = flag.Int("port", 50051, "The server port")
	promPort = flag.Int("prom-port", 9090, "Prometheus metrics port")
	env      = flag.String("env", "", "Environment: dev, beta, prod")
	dataID   = flag.String("dataid", "order", "Nacos Data ID")
)

func main() {
	// 解析命令行参数
	flag.Parse()

	// 设置环境变量（在 init 执行前设置）
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

	// 初始化 Prometheus
	metrics.InitPrometheus()
	log.Printf("[Prometheus] 指标采集器初始化成功")

	// 启动 Prometheus HTTP 服务（用于指标暴露）
	go startPrometheusServer(*promPort)

	// 启动 gRPC 服务
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 创建 gRPC 服务器（带拦截器）
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),   // 异常恢复
			interceptor.PrometheusInterceptor(), // Prometheus 监控
		),
	)

	__.RegisterOrderServiceServer(s, &service.Server{})
	log.Printf("server listening at %v", lis.Addr())
	log.Printf("Prometheus metrics available at http://localhost:%d/metrics", *promPort)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// startPrometheusServer 启动 Prometheus HTTP 服务
func startPrometheusServer(port int) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler()) // Prometheus 指标端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
<html>
<head><title>Order Service Metrics</title></head>
<body>
<h1>Order Service Metrics</h1>
<p><a href="/metrics">Prometheus Metrics</a></p>
<p><a href="/health">Health Check</a></p>
</body>
</html>
`))
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("[Prometheus] HTTP server starting at %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("[Prometheus] HTTP server error: %v", err)
	}
}
