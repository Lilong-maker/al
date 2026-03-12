package main

import (
	"al/srv/dasic/config"
	_ "al/srv/dasic/inits"
	"context"
	"fmt"
	"time"
)

func main() {
	ctx := context.Background()

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║              Nacos 配置中心 - 效果演示                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ========== 1. 显示当前环境配置 ==========
	fmt.Println("【1】当前环境配置")
	fmt.Println("─────────────────────────────────────────")
	fmt.Printf("  运行环境: %s\n", config.CurrentEnv)
	fmt.Printf("  Nacos Group: %s\n", config.Gen.Nacos.Group)
	fmt.Println()

	// ========== 2. MySQL 配置效果 ==========
	fmt.Println("【2】MySQL 配置（从 Nacos 读取）")
	fmt.Println("─────────────────────────────────────────")
	fmt.Printf("  地址: %s:%d\n", config.Gen.Mysql.Host, config.Gen.Mysql.Port)
	fmt.Printf("  数据库: %s\n", config.Gen.Mysql.Database)
	fmt.Printf("  连接池:\n")
	fmt.Printf("    - 最大连接数: %d\n", config.Gen.Mysql.MaxOpenConns)
	fmt.Printf("    - 空闲连接数: %d\n", config.Gen.Mysql.MaxIdleConns)
	fmt.Printf("    - 连接存活时间: %d 秒\n", config.Gen.Mysql.ConnMaxLifetime)
	fmt.Println()

	// ========== 3. Redis 配置效果 ==========
	fmt.Println("【3】Redis 配置（从 Nacos 读取）")
	fmt.Println("─────────────────────────────────────────")
	fmt.Printf("  地址: %s:%d\n", config.Gen.Redis.Host, config.Gen.Redis.Port)
	fmt.Printf("  连接池大小: %d\n", config.Gen.Redis.PoolSize)
	fmt.Printf("  最小空闲连接: %d\n", config.Gen.Redis.MinIdleConns)
	fmt.Println()

	// ========== 4. 测试 Redis 连接 ==========
	fmt.Println("【4】测试 Redis 连接")
	fmt.Println("─────────────────────────────────────────")

	// Ping 测试
	pong, err := config.Redis.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("  ❌ Redis 连接失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ Redis 连接成功: %s\n", pong)
	}
	fmt.Println()

	// ========== 5. 演示 Redis 缓存功能 ==========
	fmt.Println("【5】演示 Redis 缓存功能")
	fmt.Println("─────────────────────────────────────────")

	// 设置缓存
	key := "demo:user:1001"
	value := "张三"

	fmt.Printf("  → 设置缓存: %s = %s\n", key, value)
	err = config.Redis.Set(ctx, key, value, 10*time.Minute).Err()
	if err != nil {
		fmt.Printf("  ❌ 设置失败: %v\n", err)
	} else {
		fmt.Println("  ✅ 设置成功")
	}

	// 读取缓存
	fmt.Printf("  → 读取缓存: %s\n", key)
	result, err := config.Redis.Get(ctx, key).Result()
	if err != nil {
		fmt.Printf("  ❌ 读取失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 读取成功: %s\n", result)
	}

	// 设置过期时间
	fmt.Printf("  → 设置过期时间: 10分钟\n")
	ttl, _ := config.Redis.TTL(ctx, key).Result()
	fmt.Printf("  ✅ 剩余时间: %v\n", ttl)

	// 删除缓存
	fmt.Printf("  → 删除缓存: %s\n", key)
	config.Redis.Del(ctx, key)
	_, err = config.Redis.Get(ctx, key).Result()
	if err != nil {
		fmt.Println("  ✅ 删除成功（读取不到数据）")
	}
	fmt.Println()

	// ========== 6. 演示业务配置 ==========
	fmt.Println("【6】业务配置（从 Nacos 读取）")
	fmt.Println("─────────────────────────────────────────")
	fmt.Printf("  用户服务:\n")
	fmt.Printf("    - 超时时间: %d 秒\n", config.Gen.UserService.Timeout)
	fmt.Printf("    - 最大重试: %d 次\n", config.Gen.UserService.MaxRetry)
	fmt.Printf("    - 限流 QPS: %d\n", config.Gen.UserService.RateLimit)
	fmt.Printf("    - 启用缓存: %v\n", config.Gen.UserService.EnableCache)
	fmt.Printf("    - 缓存 TTL: %d 秒\n", config.Gen.UserService.CacheTTL)
	fmt.Println()

	fmt.Printf("  订单服务:\n")
	fmt.Printf("    - 超时时间: %d 秒\n", config.Gen.OrderService.Timeout)
	fmt.Printf("    - 最大重试: %d 次\n", config.Gen.OrderService.MaxRetry)
	fmt.Printf("    - 最大金额: %.2f\n", config.Gen.OrderService.MaxOrderAmount)
	fmt.Printf("    - 异步处理: %v\n", config.Gen.OrderService.EnableAsync)
	fmt.Println()

	// ========== 7. 演示热更新 ==========
	fmt.Println("【7】热更新演示")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("  请在 Nacos 控制台修改配置，观察下方变化...")
	fmt.Println()
	fmt.Println("  操作步骤:")
	fmt.Println("  1. 打开 http://115.190.43.83:8848/nacos")
	fmt.Println("  2. 找到 DataID='order', Group='dev' 的配置")
	fmt.Println("  3. 修改 UserService.Timeout 或其他字段")
	fmt.Println("  4. 点击'发布'")
	fmt.Println()
	fmt.Println("  当前 UserService.Timeout =", config.Gen.UserService.Timeout, "秒")
	fmt.Println()
	fmt.Println("  按 Ctrl+C 退出程序...")
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("等待配置变更中...（修改 Nacos 后会自动打印变更信息）")
	fmt.Println("═══════════════════════════════════════════════════════════════")

	// 持续监听，等待用户操作
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			// 每5秒打印一次当前超时时间
			fmt.Printf("[%s] 当前 Timeout = %d 秒\n",
				time.Now().Format("15:04:05"),
				config.Gen.UserService.Timeout)
		}
	}
}
