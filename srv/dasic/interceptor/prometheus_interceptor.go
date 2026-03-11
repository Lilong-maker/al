// Package interceptor gRPC 拦截器
package interceptor

import (
	"context"
	"fmt"
	"time"

	"al/srv/dasic/metrics"

	"google.golang.org/grpc"
)

// PrometheusInterceptor Prometheus 监控拦截器
// 自动采集 P99 延迟、QPS、错误率等指标
func PrometheusInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 开始时间
		start := time.Now()
		method := info.FullMethod

		// 增加正在处理的请求数
		if metrics.GlobalPrometheusAgent != nil {
			metrics.GlobalPrometheusAgent.IncInFlight(method)
		}

		// 执行请求
		resp, err := handler(ctx, req)

		// 计算耗时
		duration := time.Since(start)

		// 减少正在处理的请求数
		if metrics.GlobalPrometheusAgent != nil {
			metrics.GlobalPrometheusAgent.DecInFlight(method)
		}

		// 记录指标
		recordMetrics(method, duration, err)

		// 打印慢请求日志（超过 500ms）
		if duration > 500*time.Millisecond {
			fmt.Printf("[慢请求告警] %s 耗时 %v\n", method, duration)
		}

		return resp, err
	}
}

// recordMetrics 记录指标
func recordMetrics(method string, duration time.Duration, err error) {
	if metrics.GlobalPrometheusAgent == nil {
		return
	}

	// 判断状态
	status := "success"
	if err != nil {
		status = "error"
		// 记录错误
		metrics.GlobalPrometheusAgent.RecordError(method, "grpc_error")
	}

	// 记录请求
	metrics.GlobalPrometheusAgent.RecordRequest(method, duration, status)
}

// RecoveryInterceptor 异常恢复拦截器
// 防止 panic 导致服务崩溃
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				// 记录 panic 错误
				if metrics.GlobalPrometheusAgent != nil {
					metrics.GlobalPrometheusAgent.RecordError(info.FullMethod, "panic")
				}
				fmt.Printf("[Panic] %s 发生 panic: %v\n", info.FullMethod, r)
				err = fmt.Errorf("internal error: %v", r)
			}
		}()

		return handler(ctx, req)
	}
}

// ChainInterceptors 拦截器链
func ChainInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 构建拦截器链
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chain
			chain = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return interceptor(currentCtx, currentReq, info, next)
			}
		}
		return chain(ctx, req)
	}
}
