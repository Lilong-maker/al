package main

import (
	"al/srv/dasic/config"
	_ "al/srv/dasic/inits"
	"al/srv/handler/model"
	"fmt"
	"gospacex"
	"time"
)

func main() {

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     MySQL 三条件组合查询 - 三种方法演示                      ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 方法1：基础查询（单条件）
	method1()

	// 方法2：组合查询（三条件）
	method2()

	// 方法3：游标分页查询（优化深分页）
	method3()
}

// ========== 方法1：基础查询（单条件）==========
func method1() {
	fmt.Println("【方法1】基础查询（单条件 - 按名称模糊查询）")
	fmt.Println("─────────────────────────────────────────")

	// 查询条件：名称包含 "iPhone"
	name := "iPhone"

	var orders []model.Order
	db := config.DB.Model(&model.Order{})
	db = db.Where("name LIKE ?", name+"%") // 右模糊，可用索引

	err := db.Limit(5).Find(&orders).Error
	if err != nil {
		fmt.Printf("  ❌ 查询失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 查询成功，找到 %d 条记录\n", len(orders))
	for _, order := range orders {
		fmt.Printf("     - ID: %d, 名称: %s, 价格: %.2f\n", order.ID, order.Name, order.Price)
	}
	fmt.Println()
}

// ========== 方法2：组合查询（三条件）==========
func method2() {
	fmt.Println("【方法2】组合查询（三条件 - 名称 + 价格 + 时间）")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("  条件1: 名称 LIKE 'iPhone%'")
	fmt.Println("  条件2: 价格 BETWEEN 1000 AND 5000")
	fmt.Println("  条件3: 创建时间在最近30天")
	fmt.Println()

	// 构建查询
	db := config.DB.Model(&model.Order{})
	var conditions []string

	// 条件1：名称模糊查询
	name := "iPhone"
	db = db.Where("name LIKE ?", name+"%")
	conditions = append(conditions, "名称 LIKE '"+name+"%'")

	// 条件2：价格区间
	minPrice := 1000.0
	maxPrice := 5000.0
	db = db.Where("price >= ? AND price <= ?", minPrice, maxPrice)
	conditions = append(conditions, fmt.Sprintf("价格 %.0f-%.0f", minPrice, maxPrice))

	// 条件3：创建时间范围（最近30天）
	startTime := time.Now().AddDate(0, 0, -30).Format("2024-01-01 00:00:00")
	endTime := time.Now().Format("2024-12-31 23:59:59")
	db = db.Where("created_at >= ? AND created_at <= ?", startTime, endTime)
	conditions = append(conditions, "时间范围")

	fmt.Printf("  查询条件: %v\n", conditions)

	// 查询总数
	var total int64
	db.Count(&total)
	fmt.Printf("  总记录数: %d\n", total)

	// 分页查询
	var orders []model.Order
	err := db.Order("id DESC").Limit(5).Find(&orders).Error
	if err != nil {
		fmt.Printf("  ❌ 查询失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 查询成功，显示前 %d 条:\n", len(orders))
	for _, order := range orders {
		fmt.Printf("     - ID: %d, 名称: %s, 价格: %.2f, 数量: %d\n",
			order.ID, order.Name, order.Price, order.Num)
	}

	// 生成的 SQL 示例
	fmt.Println()
	fmt.Println("  【生成的 SQL】")
	fmt.Println("  SELECT * FROM orders")
	fmt.Println("  WHERE name LIKE 'iPhone%'")
	fmt.Println("    AND price >= 1000 AND price <= 5000")
	fmt.Println("    AND created_at >= '2024-01-01' AND created_at <= '2024-12-31'")
	fmt.Println("  ORDER BY id DESC")
	fmt.Println("  LIMIT 5")
	fmt.Println()
}

// ========== 方法3：游标分页查询（优化深分页）==========
func method3() {
	fmt.Println("【方法3】游标分页查询（优化深分页）")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("  场景：查询第 10001-10010 条记录（深分页）")
	fmt.Println()

	// ❌ 传统分页（性能差）
	fmt.Println("  ❌ 传统分页（性能差）：")
	fmt.Println("     OFFSET 10000 LIMIT 10")
	fmt.Println("     需要扫描 10010 条记录，扔掉前 10000 条")
	fmt.Println()

	// ✅ 游标分页（性能优）
	fmt.Println("  ✅ 游标分页（性能优）：")
	fmt.Println("     WHERE id < last_id LIMIT 10")
	fmt.Println("     直接定位，不需要跳过记录")
	fmt.Println()

	// 模拟：已知上一页最后一条记录的 ID 是 10000
	lastID := int64(10000)
	pageSize := 10

	db := config.DB.Model(&model.Order{})
	db = db.Where("id < ?", lastID) // 降序，所以用 <

	var orders []model.Order
	err := db.Order("id DESC").Limit(pageSize).Find(&orders).Error
	if err != nil {
		fmt.Printf("  ❌ 查询失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 游标查询成功，找到 %d 条记录\n", len(orders))
	if len(orders) > 0 {
		fmt.Printf("  本页第一条 ID: %d，最后一条 ID: %d\n",
			orders[0].ID, orders[len(orders)-1].ID)
	}

	fmt.Println()
	fmt.Println("  【三种方法对比】")
	fmt.Println("  ┌────────────────┬─────────────────┬─────────────┐")
	fmt.Println("  │    方法        │     适用场景     │    性能     │")
	fmt.Println("  ├────────────────┼─────────────────┼─────────────┤")
	fmt.Println("  │ 方法1：单条件   │ 简单查询         │ ⭐⭐⭐      │")
	fmt.Println("  │ 方法2：三条件   │ 复杂筛选         │ ⭐⭐⭐      │")
	fmt.Println("  │ 方法3：游标分页 │ 深分页、大数据量 │ ⭐⭐⭐⭐⭐  │")
	fmt.Println("  └────────────────┴─────────────────┴─────────────┘")
	fmt.Println()
}

// ========== 索引优化建议 ==========
func indexSuggestion() {
	fmt.Println("【索引优化建议】")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println()
	fmt.Println("在 MySQL 中执行以下 SQL 创建索引：")
	fmt.Println()
	fmt.Println("-- 1. 单列索引")
	fmt.Println("CREATE INDEX idx_name ON orders(name);")
	fmt.Println("CREATE INDEX idx_price ON orders(price);")
	fmt.Println("CREATE INDEX idx_created_at ON orders(created_at);")
	fmt.Println()
	fmt.Println("-- 2. 组合索引（推荐，最左前缀原则）")
	fmt.Println("CREATE INDEX idx_name_price_time ON orders(name, price, created_at);")
	fmt.Println()
	fmt.Println("-- 3. 覆盖索引（查询的字段都在索引中）")
	fmt.Println("CREATE INDEX idx_name_price_cover ON orders(name, price, id);")
	fmt.Println()
}
