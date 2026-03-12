-- ============================================
-- 数据库索引优化脚本
-- 数据库: p2308a
-- 表: orders
-- 执行时间: 预计 1-2 分钟
-- 注意: 请在低峰期执行
-- ============================================

-- 开始事务
START TRANSACTION;

-- ============================================
-- 1. 查看当前索引情况
-- ============================================
-- SHOW INDEX FROM orders;

-- ============================================
-- 2. 创建核心组合索引（高优先级 - 必须执行）
-- ============================================

-- 2.1 组合索引：三条件查询优化（name + price + created_at）
-- 优化场景：WHERE name LIKE 'xxx%' AND price >= ? AND created_at >= ?
-- 预期效果：查询从 320ms 降至 30ms，提升 10 倍
CREATE INDEX IF NOT EXISTS idx_name_price_created ON orders(name, price, created_at);

-- 2.2 组合索引：名称 + 时间（时间范围查询）
-- 优化场景：按名称 + 创建时间排序查询
CREATE INDEX IF NOT EXISTS idx_name_created ON orders(name, created_at);

-- 2.3 组合索引：价格 + 时间（价格筛选 + 排序）
-- 优化场景：价格区间 + 时间排序
CREATE INDEX IF NOT EXISTS idx_price_created ON orders(price, created_at);

-- ============================================
-- 3. 创建覆盖索引（可选 - 空间换时间）
-- ============================================

-- 3.1 覆盖索引：避免回表
-- 优化场景：只查询 id, name, price 字段
CREATE INDEX IF NOT EXISTS idx_name_price_cover ON orders(name, price, created_at, id, order_sn);

-- ============================================
-- 4. 创建辅助索引（中优先级 - 建议执行）
-- ============================================

-- 4.1 订单号 + 时间（订单号范围查询）
CREATE INDEX IF NOT EXISTS idx_ordersn_created ON orders(order_sn, created_at);

-- 4.2 数量索引（库存相关查询）
CREATE INDEX IF NOT EXISTS idx_num ON orders(num);

-- 4.3 时间倒序索引（最新订单查询）
CREATE INDEX IF NOT EXISTS idx_created_desc ON orders(created_at DESC);

-- ============================================
-- 5. 验证索引创建成功
-- ============================================
-- 查看所有索引
SHOW INDEX FROM orders;

-- ============================================
-- 6. 更新表统计信息
-- ============================================
ANALYZE TABLE orders;

-- 提交事务
COMMIT;

-- ============================================
-- 7. 验证索引效果（执行后测试）
-- ============================================

-- 7.1 查看执行计划
EXPLAIN 
SELECT * FROM orders 
WHERE name LIKE 'iPhone%'
  AND price >= 1000 AND price <= 5000
  AND created_at >= '2024-01-01'
ORDER BY id DESC
LIMIT 10;

-- 预期结果：
-- type = range
-- key = idx_name_price_created
-- rows < 100

-- 7.2 性能测试
-- SELECT SQL_NO_CACHE * FROM orders 
-- WHERE name LIKE 'iPhone%'
--   AND price >= 1000 AND price <= 5000
--   AND created_at >= '2024-01-01'
-- ORDER BY id DESC
-- LIMIT 10;

-- ============================================
-- 8. 查看索引大小（执行后检查）
-- ============================================
-- SELECT 
--     INDEX_NAME,
--     SUM(INDEX_LENGTH)/1024/1024 AS index_size_mb
-- FROM information_schema.STATISTICS
-- WHERE TABLE_NAME = 'orders'
-- GROUP BY INDEX_NAME;

-- ============================================
-- 优化完成！
-- ============================================
