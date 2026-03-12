// Package interceptor gRPC 拦截器
package interceptor

import (
	"context"
	"fmt"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SentinelInterceptor Sentinel 限流熔断拦截器
// 参考 Gin 中间件实现，适配 gRPC interceptor
func SentinelInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 资源名：使用 gRPC 方法名
		resourceName := info.FullMethod

		// 进入 Sentinel 保护
		entry, blockError := sentinel.Entry(resourceName)

		// 被限流或熔断
		if blockError != nil {
			// 返回 gRPC 标准错误
			return nil, status.Errorf(
				codes.ResourceExhausted,
				"服务繁忙，请稍后重试 (资源: %s)", resourceName,
			)
		}

		// 确保退出 entry
		defer entry.Exit()

		// 执行实际请求
		resp, err := handler(ctx, req)

		// 记录结果（用于熔断统计）
		if err != nil {
			// 记录错误
			sentinel.TraceError(entry, err)
		}

		return resp, err
	}
}

// SentinelStreamInterceptor Sentinel 流式请求拦截器
func SentinelStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		resourceName := info.FullMethod

		entry, blockError := sentinel.Entry(resourceName)

		if blockError != nil {
			return status.Errorf(
				codes.ResourceExhausted,
				"服务繁忙，请稍后重试 (资源: %s)", resourceName,
			)
		}

		defer entry.Exit()

		err := handler(srv, ss)
		if err != nil {
			sentinel.TraceError(entry, err)
		}

		return err
	}
}

// SentinelInterceptorWithFallback Sentinel 限流拦截器（带降级函数）
func SentinelInterceptorWithFallback(fallbackHandler grpc.UnaryHandler) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resourceName := info.FullMethod

		entry, blockError := sentinel.Entry(resourceName)

		if blockError != nil {
			// 被限流时执行降级逻辑
			if fallbackHandler != nil {
				fmt.Printf("[Sentinel] 触发降级: %s\n", resourceName)
				return fallbackHandler(ctx, req)
			}
			return nil, status.Errorf(
				codes.ResourceExhausted,
				"服务繁忙，请稍后重试 (资源: %s)", resourceName,
			)
		}

		defer entry.Exit()

		resp, err := handler(ctx, req)
		if err != nil {
			sentinel.TraceError(entry, err)
		}

		return resp, err
	}
}
