package main

import (
	"al/srv/dasic/config"
	_ "al/srv/dasic/init"
	"al/srv/handler/service"
	"fmt"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("热配置演示程序")
	fmt.Println("========================================")
	fmt.Println()

	// 显示当前配置
	fmt.Println("【当前配置】从 Nacos 读取到的配置：")
	fmt.Printf("  UserService.Timeout: %d 秒\n", config.Gen.UserService.Timeout)
	fmt.Printf("  UserService.MaxRetry: %d 次\n", config.Gen.UserService.MaxRetry)
	fmt.Printf("  UserService.RateLimit: %d QPS\n", config.Gen.UserService.RateLimit)
	fmt.Printf("  UserService.EnableCache: %v\n", config.Gen.UserService.EnableCache)
	fmt.Printf("  UserService.CacheTTL: %d 秒\n", config.Gen.UserService.CacheTTL)
	fmt.Println()

	// 模拟用户请求
	fmt.Println("【模拟请求】使用当前配置处理请求：")
	service.GlobalHotConfigDemo.SimulateUserRequest(1001)
	service.GlobalHotConfigDemo.SimulateUserRequest(1002)
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("现在请在 Nacos 控制台修改配置，然后按 Enter 键继续...")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("操作步骤：")
	fmt.Println("1. 打开浏览器，访问 Nacos 控制台")
	fmt.Println("2. 找到 DataID='order', Group='dev' 的配置")
	fmt.Println("3. 修改以下字段：")
	fmt.Println("   - UserService.Timeout: 30 -> 10")
	fmt.Println("   - UserService.MaxRetry: 3 -> 5")
	fmt.Println("   - UserService.EnableCache: true -> false")
	fmt.Println("4. 点击'发布'")
	fmt.Println("5. 回到这里按 Enter 键")
	fmt.Scanln()

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("【配置更新后】从 Nacos 读取到的最新配置：")
	fmt.Printf("  UserService.Timeout: %d 秒\n", config.Gen.UserService.Timeout)
	fmt.Printf("  UserService.MaxRetry: %d 次\n", config.Gen.UserService.MaxRetry)
	fmt.Printf("  UserService.RateLimit: %d QPS\n", config.Gen.UserService.RateLimit)
	fmt.Printf("  UserService.EnableCache: %v\n", config.Gen.UserService.EnableCache)
	fmt.Printf("  UserService.CacheTTL: %d 秒\n", config.Gen.UserService.CacheTTL)
	fmt.Println()

	// 再次模拟用户请求
	fmt.Println("【模拟请求】使用新配置处理请求（无需重启服务）：")
	service.GlobalHotConfigDemo.SimulateUserRequest(1003)
	service.GlobalHotConfigDemo.SimulateUserRequest(1004)
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("演示完成！")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("你可以看到：")
	fmt.Println("✓ 服务没有重启")
	fmt.Println("✓ 配置自动更新了")
	fmt.Println("✓ 新请求使用了新配置")
	fmt.Println()
	fmt.Println("按 Enter 键退出...")
	fmt.Scanln()
}
