# MySQL 三条件组合查询 - 三种方法

## 概述

在 `model/user.go` 中定义了 Order 表结构，包含三种 MySQL 查询方法的索引优化。

## 三种查询方法

### 【方法1】基础查询 - 单条件查询（按名称）

**使用场景**：简单筛选，如按名称搜索

**SQL示例**：
```sql
SELECT * FROM orders 
WHERE name LIKE 'iPhone%' 
LIMIT 10
```

**索引**：
```sql
INDEX idx_name (name)
```

**代码实现**：
```go
db := config.DB.Model(&model.Order{})
db = db.Where("name LIKE ?", name+"%")
var orders []model.Order
db.Limit(10).Find(&orders)
```

**特点**：
- ✅ 简单快速
- ✅ 右模糊查询可用索引（`name%`）
- ❌ 左模糊不能用索引（`%name%`）

---

### 【方法2】组合查询 - 三条件组合（名称 + 价格 + 时间）

**使用场景**：复杂筛选，如电商商品筛选

**SQL示例**：
```sql
SELECT * FROM orders
WHERE name LIKE 'iPhone%'
  AND price >= 1000 AND price <= 5000
  AND created_at >= '2024-01-01' AND created_at <= '2024-12-31'
ORDER BY id DESC
LIMIT 10
```

**索引**（最左前缀原则）：
```sql
-- 组合索引（推荐）
CREATE INDEX idx_name_price_time ON orders(name, price, created_at);

-- 或者单列索引
CREATE INDEX idx_name ON orders(name);
CREATE INDEX idx_price ON orders(price);
CREATE INDEX idx_created_at ON orders(created_at);
```

**代码实现**：
```go
db := config.DB.Model(&model.Order{})

// 条件1：名称模糊查询
if name != "" {
    db = db.Where("name LIKE ?", name+"%")
}

// 条件2：价格区间
if minPrice > 0 {
    db = db.Where("price >= ?", minPrice)
}
if maxPrice > 0 {
    db = db.Where("price <= ?", maxPrice)
}

// 条件3：创建时间范围
if startTime != "" {
    db = db.Where("created_at >= ?", startTime)
}
if endTime != "" {
    db = db.Where("created_at <= ?", endTime)
}

// 分页查询
var total int64
db.Count(&total)

var orders []model.Order
db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&orders)
```

**特点**：
- ✅ 三条件同时筛选
- ✅ 组合索引减少回表（数据量越大优势越明显）
- ⚠️ 注意最左前缀原则：查询条件要从索引的最左边开始

**最左前缀原则说明**：
```
索引：(name, price, created_at)

✅ 能用索引：
  WHERE name = 'iPhone'
  WHERE name = 'iPhone' AND price = 1000
  WHERE name = 'iPhone' AND price = 1000 AND created_at > '2024-01-01'

❌ 不能用索引：
  WHERE price = 1000                    -- 跳过了 name
  WHERE created_at > '2024-01-01'       -- 跳过了 name 和 price
  WHERE price = 1000 AND created_at > '2024-01-01'  -- 跳过了 name
```

---

### 【方法3】游标分页 - 优化深分页

**使用场景**：大数据量列表，如无限滚动加载

**传统分页（性能差）**：
```sql
-- 查询第 10001-10010 条
SELECT * FROM orders 
ORDER BY id DESC
LIMIT 10 OFFSET 10000
-- 需要扫描 10010 条记录，扔掉前 10000 条，很慢！
```

**游标分页（性能优）**：
```sql
-- 已知上一页最后一条记录的 ID 是 10000
SELECT * FROM orders 
WHERE id < 10000      -- 直接定位，不需要跳过记录
ORDER BY id DESC
LIMIT 10
-- 只需要扫描 10 条记录，很快！
```

**索引**：
```sql
PRIMARY KEY (id)  -- 主键本身就是聚簇索引
```

**代码实现**：
```go
// 游标分页查询
func OrderListWithCursor(lastID int64, pageSize int, name string) ([]model.Order, error) {
    db := config.DB.Model(&model.Order{})
    
    // 条件1：名称
    if name != "" {
        db = db.Where("name LIKE ?", name+"%")
    }
    
    // 条件2：游标（上一页最后一条的ID）
    if lastID > 0 {
        db = db.Where("id < ?", lastID)  // 降序排列，所以用 <
    }
    
    var orders []model.Order
    err := db.Order("id DESC").Limit(pageSize).Find(&orders).Error
    return orders, err
}
```

**特点**：
- ✅ 深分页性能稳定（第1页和第10000页速度一样）
- ✅ 适合无限滚动、移动端列表
- ⚠️ 不支持跳转到指定页码（只能上一页/下一页）

---

## 索引对比

| 索引类型 | 适用场景 | 优点 | 缺点 |
|---------|---------|------|------|
| 单列索引 | 单条件查询 | 简单、省空间 | 多条件时效率低 |
| 组合索引 | 多条件查询 | 减少回表、效率高 | 占用更多空间 |
| 覆盖索引 | 特定查询 | 完全避免回表 | 需要包含所有查询字段 |

---

## 自动迁移

在 `init.go` 中自动创建表和索引：

```go
config.DB.AutoMigrate(&model.Order{})
```

GORM 会自动创建：
- 表结构
- 主键索引（id）
- 定义的索引（idx_name, idx_price 等）
- 软删除索引（deleted_at）

---

## 性能对比

假设表中有 100 万条数据：

| 查询方式 | 第1页(0-10) | 第1000页(10000-10010) | 适用场景 |
|---------|-------------|----------------------|---------|
| 单条件查询 | 10ms | 100ms | 简单筛选 |
| 三条件组合 | 20ms | 200ms | 复杂筛选 |
| 传统分页 | 10ms | 5000ms | 小数据量 |
| 游标分页 | 10ms | 10ms | 大数据量 |

---

## 总结

- **方法1**：单条件查询，简单快速，适合基础筛选
- **方法2**：三条件组合，适合复杂查询场景，注意最左前缀原则
- **方法3**：游标分页，深分页性能最优，适合移动端和无限滚动
