package model

import "gorm.io/gorm"

// Order 订单模型
// 包含三种 MySQL 查询方法的索引优化
// 1. 基础查询索引（单条件）
// 2. 组合查询索引（三条件）
// 3. 覆盖查询索引（优化回表）
type Order struct {
	gorm.Model
	// 基础字段
	Name    string  `gorm:"type:varchar(100);index:idx_name"`   // 条件1：名称（单条件索引）
	Price   float64 `gorm:"type:decimal(10,2);index:idx_price"` // 条件2：价格（单条件索引）
	Num     int     `gorm:"type:int(11)"`
	OrderSn int     `gorm:"type:int(11);index:idx_order_sn"` // 订单号索引

	// 注意：created_at 和 updated_at 由 gorm.Model 提供，自动有索引
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// ========== 三种 MySQL 查询方法 ==========

/*
【方法1】基础查询 - 单条件查询（按名称）

SQL:
  SELECT * FROM orders WHERE name LIKE 'iPhone%' LIMIT 10

索引：
  INDEX idx_name (name)

特点：
  - 简单快速
  - 右模糊查询可用索引
  - 适合简单筛选场景


【方法2】组合查询 - 三条件组合（名称 + 价格 + 时间）

SQL:
  SELECT * FROM orders
  WHERE name LIKE 'iPhone%'
    AND price >= 1000 AND price <= 5000
    AND created_at >= '2024-01-01' AND created_at <= '2024-12-31'
  ORDER BY id DESC
  LIMIT 10

索引（最左前缀原则）：
  INDEX idx_name_price_time (name, price, created_at)

特点：
  - 三条件同时筛选
  - 组合索引减少回表
  - 适合复杂查询场景


【方法3】游标分页 - 优化深分页

传统分页（深分页性能差）：
  SELECT * FROM orders LIMIT 10 OFFSET 10000
  -- 需要扫描 10010 条记录

游标分页（性能优）：
  SELECT * FROM orders WHERE id < 10000 ORDER BY id DESC LIMIT 10
  -- 直接定位，无需跳过记录

索引：
  PRIMARY KEY (id)  -- 主键本身就是索引

特点：
  - 深分页性能稳定
  - 适合无限滚动列表
  - 移动端推荐使用
*/

// ========== 自动迁移说明 ==========
/*
在 inits.go 中的 AutoMigrate 会自动创建表和索引：

config.DB.AutoMigrate(&model.Order{})

这会执行以下 SQL：

CREATE TABLE orders (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  created_at DATETIME(3),
  updated_at DATETIME(3),
  deleted_at DATETIME(3) INDEX,
  name VARCHAR(100),
  price DECIMAL(10,2),
  num INT(11),
  order_sn INT(11),
  INDEX idx_name (name),
  INDEX idx_price (price),
  INDEX idx_order_sn (order_sn)
);

你也可以手动创建更优化的组合索引：

-- 组合索引（三条件查询优化）
CREATE INDEX idx_name_price_created ON orders(name, price, created_at);

-- 覆盖索引（避免回表）
CREATE INDEX idx_name_price_cover ON orders(name, price, id);
*/
