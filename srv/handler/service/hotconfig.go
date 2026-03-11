package service

import (
	__ "al/proto"
	"al/srv/dasic/config"
	"context"
	"fmt"
	"sync"
	"time"
)

// HotConfigDemo 热配置演示
// 展示如何在业务代码中使用热配置
type HotConfigDemo struct {
	// 当前超时设置（会从配置中读取）
	currentTimeout int
	// 当前最大重试次数
	currentMaxRetry int
	// 是否启用缓存
	cacheEnabled bool
	// 配置变更锁
	configLock sync.RWMutex
}

// GlobalHotConfigDemo 全局热配置演示实例
var GlobalHotConfigDemo *HotConfigDemo

// InitHotConfigDemo 初始化热配置演示
func InitHotConfigDemo() {
	GlobalHotConfigDemo = NewHotConfigDemo()

	// 注册为配置观察者
	config.GlobalConfigManager.RegisterObserver(GlobalHotConfigDemo)

	fmt.Println("热配置演示服务已初始化并注册为观察者")
}

// NewHotConfigDemo 创建热配置演示实例
func NewHotConfigDemo() *HotConfigDemo {
	demo := &HotConfigDemo{
		currentTimeout:  config.Gen.UserService.Timeout,
		currentMaxRetry: config.Gen.UserService.MaxRetry,
		cacheEnabled:    config.Gen.UserService.EnableCache,
	}

	fmt.Printf("HotConfigDemo 初始化: Timeout=%d, MaxRetry=%d, EnableCache=%v\n",
		demo.currentTimeout, demo.currentMaxRetry, demo.cacheEnabled)

	return demo
}

// OnConfigChange 实现 ConfigObserver 接口
// 当配置发生变化时，这个方法会被自动调用
func (d *HotConfigDemo) OnConfigChange(changedFields []string) {
	fmt.Println("\n[HotConfigDemo] 收到配置变更通知:")

	for _, field := range changedFields {
		switch field {
		case "UserService.Timeout":
			d.updateTimeout()
		case "UserService.MaxRetry":
			d.updateMaxRetry()
		case "UserService.EnableCache":
			d.updateCacheSetting()
		case "UserService.RateLimit":
			d.updateRateLimit()
		default:
			fmt.Printf("  - 忽略未处理的字段: %s\n", field)
		}
	}
}

// updateTimeout 更新超时时间
func (d *HotConfigDemo) updateTimeout() {
	oldTimeout := d.currentTimeout
	newTimeout := config.Gen.UserService.Timeout

	d.configLock.Lock()
	d.currentTimeout = newTimeout
	d.configLock.Unlock()

	fmt.Printf("  [热更新] UserService.Timeout: %d -> %d\n", oldTimeout, newTimeout)
	fmt.Println("    下次请求将使用新的超时时间，无需重启服务！")
}

// updateMaxRetry 更新最大重试次数
func (d *HotConfigDemo) updateMaxRetry() {
	oldRetry := d.currentMaxRetry
	newRetry := config.Gen.UserService.MaxRetry

	d.configLock.Lock()
	d.currentMaxRetry = newRetry
	d.configLock.Unlock()

	fmt.Printf("  [热更新] UserService.MaxRetry: %d -> %d\n", oldRetry, newRetry)
	fmt.Println("    下次失败请求将使用新的重试次数！")
}

// updateCacheSetting 更新缓存设置
func (d *HotConfigDemo) updateCacheSetting() {
	oldEnabled := d.cacheEnabled
	newEnabled := config.Gen.UserService.EnableCache

	d.configLock.Lock()
	d.cacheEnabled = newEnabled
	d.configLock.Unlock()

	fmt.Printf("  [热更新] UserService.EnableCache: %v -> %v\n", oldEnabled, newEnabled)
	if newEnabled {
		fmt.Println("    缓存已启用，新请求将使用缓存！")
	} else {
		fmt.Println("    缓存已禁用，新请求将直接查询数据库！")
	}
}

// updateRateLimit 更新限流设置
func (d *HotConfigDemo) updateRateLimit() {
	rateLimit := config.Gen.UserService.RateLimit
	fmt.Printf("  [热更新] UserService.RateLimit -> %d\n", rateLimit)
	fmt.Println("    限流阈值已更新，新请求将应用新的限流规则！")
}

// GetCurrentConfig 获取当前配置（带锁）
func (d *HotConfigDemo) GetCurrentConfig() (timeout, maxRetry int, cacheEnabled bool) {
	d.configLock.RLock()
	defer d.configLock.RUnlock()

	return d.currentTimeout, d.currentMaxRetry, d.cacheEnabled
}

// SimulateUserRequest 模拟用户请求（演示热配置的使用）
func (d *HotConfigDemo) SimulateUserRequest(userID int64) {
	timeout, maxRetry, cacheEnabled := d.GetCurrentConfig()

	fmt.Printf("\n[模拟请求] 用户 %d:\n", userID)
	fmt.Printf("  - 使用当前超时: %d 秒\n", timeout)
	fmt.Printf("  - 使用最大重试: %d 次\n", maxRetry)
	fmt.Printf("  - 缓存状态: %v\n", cacheEnabled)

	// 模拟实际业务逻辑...
	if cacheEnabled {
		fmt.Println("  -> 尝试从缓存获取数据...")
	} else {
		fmt.Println("  -> 直接查询数据库...")
	}
}

// ============ 在 gRPC Service 中使用热配置的示例 ============

// ServerWithHotConfig 展示如何在 gRPC 服务中使用热配置
type ServerWithHotConfig struct {
	__   *__.UnimplementedOrderServiceServer
	demo *HotConfigDemo
}

// NewServerWithHotConfig 创建支持热配置的服务器
func NewServerWithHotConfig() *ServerWithHotConfig {
	return &ServerWithHotConfig{
		demo: GlobalHotConfigDemo,
	}
}

// ProcessOrderWithTimeout 处理订单（使用热配置的超时设置）
func (s *ServerWithHotConfig) ProcessOrderWithTimeout(ctx context.Context, orderID int64) error {
	// 获取当前热配置的超时时间
	timeout, maxRetry, _ := s.demo.GetCurrentConfig()

	fmt.Printf("\n[业务处理] 订单 %d:\n", orderID)
	fmt.Printf("  - 当前超时设置: %d 秒（从 Nacos 热配置读取）\n", timeout)
	fmt.Printf("  - 当前重试设置: %d 次（从 Nacos 热配置读取）\n", maxRetry)

	// 使用配置创建带超时的 context
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 模拟业务处理
	select {
	case <-ctx.Done():
		fmt.Println("  - 处理超时！")
		return fmt.Errorf("订单处理超时")
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  - 处理完成！")
		return nil
	}
}
