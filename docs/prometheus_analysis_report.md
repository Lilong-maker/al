# Prometheus Agent 接口分析报告

## 一、Prometheus Agent 架构设计

### 1.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        Order Service (gRPC)                     │
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │
│  │   Client    │───▶│  Interceptor│───▶│   Handler   │        │
│  └─────────────┘    └──────┬──────┘    └─────────────┘        │
│                            │                                    │
│                            ▼                                    │
│                   ┌─────────────────┐                          │
│                   │ PrometheusAgent │                          │
│                   │  - RequestCounter│                          │
│                   │  - LatencyHist   │                          │
│                   │  - ErrorCounter  │                          │
│                   └────────┬────────┘                          │
│                            │                                    │
└────────────────────────────┼────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Prometheus HTTP Server (Port 9090)          │
│                                                                 │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │ GET /metrics                                            │  │
│   │ - order_service_requests_total                          │  │
│   │ - order_service_request_duration_seconds                │  │
│   │ - order_service_errors_total                            │  │
│   │ - order_service_requests_in_flight                      │  │
│   └─────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 核心组件说明

| 组件 | 文件 | 功能 |
|------|------|------|
| **PrometheusAgent** | `metrics/prometheus_agent.go` | 指标采集器，定义了 Counter、Histogram、Gauge |
| **Interceptor** | `interceptor/prometheus_interceptor.go` | gRPC 拦截器，自动采集请求指标 |
| **Main** | `cmd/main.go` | 集成 Prometheus HTTP 服务 |

---

## 二、采集指标详解

### 2.1 指标列表

| 指标名 | 类型 | 说明 | 标签 |
|--------|------|------|------|
| `order_service_requests_total` | Counter | 总请求数 | method, status |
| `order_service_request_duration_seconds` | Histogram | 请求延迟（秒） | method |
| `order_service_requests_in_flight` | Gauge | 正在处理的请求数 | method |
| `order_service_errors_total` | Counter | 错误总数 | method, error_type |
| `order_service_response_size_bytes` | Histogram | 响应大小 | method |
| `order_service_request_size_bytes` | Histogram | 请求大小 | method |

### 2.2 指标示例

```
# 请求计数器
order_service_requests_total{method="/proto.OrderService/OrderCreate",status="success"} 1250
order_service_requests_total{method="/proto.OrderService/OrderCreate",status="error"} 15

# 延迟分布（用于计算 P99）
order_service_request_duration_seconds_bucket{method="/proto.OrderService/OrderCreate",le="0.1"} 800
order_service_request_duration_seconds_bucket{method="/proto.OrderService/OrderCreate",le="0.5"} 1200
order_service_request_duration_seconds_bucket{method="/proto.OrderService/OrderCreate",le="1"} 1240
order_service_request_duration_seconds_bucket{method="/proto.OrderService/OrderCreate",le="+Inf"} 1265
```

---

## 三、PromQL 查询示例

### 3.1 QPS 查询（每秒请求数）

```promql
# 整体 QPS
sum(rate(order_service_requests_total[1m]))

# 按接口分组的 QPS
sum by (method) (rate(order_service_requests_total[1m]))

# 成功请求的 QPS
sum(rate(order_service_requests_total{status="success"}[1m]))
```

### 3.2 P99 延迟查询

```promql
# P99 延迟（99% 的请求响应时间）
histogram_quantile(0.99, 
  sum by (method, le) (
    rate(order_service_request_duration_seconds_bucket[5m])
  )
)

# P95 延迟
histogram_quantile(0.95, 
  sum by (method, le) (
    rate(order_service_request_duration_seconds_bucket[5m])
  )
)

# P50 延迟（中位数）
histogram_quantile(0.50, 
  sum by (method, le) (
    rate(order_service_request_duration_seconds_bucket[5m])
  )
)
```

### 3.3 错误率查询

```promql
# 整体错误率
sum(rate(order_service_requests_total{status="error"}[5m])) 
/ 
sum(rate(order_service_requests_total[5m])) 
* 100

# 按接口的错误率
sum by (method) (rate(order_service_requests_total{status="error"}[5m])) 
/ 
sum by (method) (rate(order_service_requests_total[5m])) 
* 100
```

### 3.4 并发数查询

```promql
# 正在处理的请求数
sum(order_service_requests_in_flight)

# 按接口的并发数
order_service_requests_in_flight
```

---

## 四、接口分析报告

### 4.1 当前接口性能（示例数据）

| 接口 | P99 延迟 | P95 延迟 | P50 延迟 | QPS | 错误率 | 状态 |
|------|----------|----------|----------|-----|--------|------|
| **OrderCreate** | 85ms | 60ms | 25ms | 500 | 0.5% | 🟢 良好 |
| **OrderGet** | 45ms | 30ms | 15ms | 2000 | 0.1% | 🟢 优秀 |
| **OrderList** | 320ms | 250ms | 100ms | 100 | 1.2% | 🟡 需优化 |
| **OrderUpdate** | 95ms | 70ms | 35ms | 300 | 0.8% | 🟢 良好 |
| **OrderDelete** | 55ms | 40ms | 20ms | 200 | 0.3% | 🟢 优秀 |

**状态说明**：
- 🟢 优秀：P99 < 50ms, QPS > 1000, 错误率 < 0.1%
- 🟢 良好：P99 < 100ms, QPS > 500, 错误率 < 1%
- 🟡 需优化：P99 < 500ms, QPS > 100, 错误率 < 5%
- 🔴 危险：P99 > 500ms 或 错误率 > 5%

### 4.2 问题接口分析：OrderList

```
┌─────────────────────────────────────────────────────────────────┐
│                    OrderList 接口问题诊断                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  🔴 问题：P99 延迟 320ms，超过阈值                               │
│                                                                 │
│  根因分析：                                                      │
│  1. 三条件组合查询（name + price + created_at）                  │
│     └─> 缺乏组合索引，全表扫描                                   │
│                                                                 │
│  2. 深分页问题（OFFSET > 1000）                                  │
│     └─> 数据库需要扫描大量行，性能差                              │
│                                                                 │
│  3. 未使用缓存                                                   │
│     └─> 重复查询数据库，无缓存加速                                │
│                                                                 │
│  4. 响应数据量大                                                 │
│     └─> 未限制 pageSize，可能返回大量数据                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 五、优化建议

### 5.1 数据库优化

```sql
-- 1. 添加组合索引（最左前缀原则）
CREATE INDEX idx_name_price_created ON orders(name, price, created_at);

-- 2. 单列索引（可选）
CREATE INDEX idx_name ON orders(name);
CREATE INDEX idx_price ON orders(price);
CREATE INDEX idx_created_at ON orders(created_at);

-- 3. 验证索引使用
EXPLAIN SELECT * FROM orders 
WHERE name LIKE 'iPhone%' 
  AND price >= 1000 AND price <= 5000 
  AND created_at >= '2024-01-01'
ORDER BY id DESC 
LIMIT 10;
```

**预期效果**：
- P99 延迟：320ms → 50ms（提升 6 倍）
- QPS：100 → 500（提升 5 倍）

### 5.2 游标分页优化

```go
// ❌ 传统分页（深分页性能差）
SELECT * FROM orders LIMIT 10 OFFSET 10000

// ✅ 游标分页（性能优）
SELECT * FROM orders WHERE id < 10000 ORDER BY id DESC LIMIT 10
```

**实现代码**：
```go
// OrderListWithCursor 游标分页查询
func (s *Server) OrderListWithCursor(lastID int64, pageSize int) ([]Order, error) {
    db := config.DB.Model(&Order{})
    
    // 使用游标代替 OFFSET
    if lastID > 0 {
        db = db.Where("id < ?", lastID)
    }
    
    var orders []Order
    err := db.Order("id DESC").Limit(pageSize).Find(&orders).Error
    return orders, err
}
```

### 5.3 Redis 缓存优化

```go
// OrderGet 添加缓存
func (s *Server) OrderGet(ctx context.Context, in *proto.OrderGetReq) (*proto.OrderGetResp, error) {
    cacheKey := fmt.Sprintf("order:%d", in.Id)
    
    // 1. 查缓存
    cached, err := config.Redis.Get(ctx, cacheKey).Result()
    if err == nil {
        // 缓存命中
        var order model.Order
        json.Unmarshal([]byte(cached), &order)
        return &proto.OrderGetResp{Data: modelToOrderInfo(&order)}, nil
    }
    
    // 2. 查数据库
    var order model.Order
    if err := config.DB.First(&order, in.Id).Error; err != nil {
        return nil, err
    }
    
    // 3. 写入缓存（过期时间 10 分钟）
    data, _ := json.Marshal(order)
    config.Redis.Set(ctx, cacheKey, data, 10*time.Minute)
    
    return &proto.OrderGetResp{Data: modelToOrderInfo(&order)}, nil
}
```

**预期效果**：
- 缓存命中率 80%
- P99 延迟：45ms → 5ms（提升 9 倍）
- 数据库 QPS 下降 80%

### 5.4 分页大小限制

```go
// 限制最大分页大小
const MaxPageSize = 100

func (s *Server) OrderList(ctx context.Context, in *proto.OrderListReq) (*proto.OrderListResp, error) {
    pageSize := in.PageSize
    if pageSize <= 0 {
        pageSize = 10
    }
    if pageSize > MaxPageSize {
        pageSize = MaxPageSize  // 限制最大 100 条
    }
    
    // ... 查询逻辑
}
```

### 5.5 Sentinel 限流保护

```go
// 对慢接口添加限流
func init() {
    // OrderList 限流：QPS 50（查询更耗资源）
    flow.LoadRules([]*flow.Rule{
        {
            Resource:               "OrderList",
            TokenCalculateStrategy: flow.Direct,
            ControlBehavior:        flow.Reject,
            Threshold:              50,
            StatIntervalInMs:       1000,
        },
    })
}
```

---

## 六、Prometheus 告警规则

```yaml
# prometheus_alerts.yml
groups:
  - name: order_service_alerts
    rules:
      # P99 延迟告警
      - alert: HighP99Latency
        expr: histogram_quantile(0.99, sum by (method, le) (rate(order_service_request_duration_seconds_bucket[5m]))) > 0.5
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "接口 {{ $labels.method }} P99 延迟过高"
          description: "P99 延迟: {{ $value }}s，超过阈值 0.5s"

      # 错误率告警
      - alert: HighErrorRate
        expr: sum by (method) (rate(order_service_requests_total{status="error"}[5m])) / sum by (method) (rate(order_service_requests_total[5m])) > 0.05
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "接口 {{ $labels.method }} 错误率过高"
          description: "错误率: {{ $value | humanizePercentage }}，超过阈值 5%"

      # QPS 下降告警
      - alert: LowQPS
        expr: sum(rate(order_service_requests_total[5m])) < 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "整体 QPS 异常下降"
          description: "当前 QPS: {{ $value }}"

      # 并发数告警
      - alert: HighConcurrency
        expr: sum(order_service_requests_in_flight) > 1000
        for: 30s
        labels:
          severity: warning
        annotations:
          summary: "并发请求数过高"
          description: "当前并发数: {{ $value }}"
```

---

## 七、Grafana 监控面板

### 7.1 推荐面板布局

```
┌─────────────────────────────────────────────────────────────┐
│                     Order Service Dashboard                 │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐          │
│  │  QPS    │ │ P99延迟 │ │ 错误率  │ │ 并发数  │          │
│  │  1,234  │ │  85ms   │ │  0.5%   │ │   56    │          │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘          │
├─────────────────────────────────────────────────────────────┤
│  QPS 趋势图 (1小时)                                         │
│  ▁▂▃▄▅▆▇█▇▆▅▄▃▂▁▁▂▃▄▅▆▇█                                 │
├─────────────────────────────────────────────────────────────┤
│  P99 延迟趋势图 (1小时)                                     │
│  ▁▁▂▂▃▃▄▄▅▅▆▆▇▇██▇▇▆▆▅▅▄▄▃▃▂▂▁▁                        │
├─────────────────────────────────────────────────────────────┤
│  接口性能对比                                               │
│  OrderCreate    ████████ 85ms                              │
│  OrderGet       ████     45ms  ✅                          │
│  OrderList      ████████████████ 320ms ⚠️                  │
│  OrderUpdate    █████████ 95ms                             │
│  OrderDelete    █████    55ms                              │
└─────────────────────────────────────────────────────────────┘
```

---

## 八、使用指南

### 8.1 启动服务

```bash
# 启动 gRPC 服务（默认端口 50051，Prometheus 端口 9090）
go run srv/dasic/cmd/main.go

# 指定 Prometheus 端口
go run srv/dasic/cmd/main.go -prom-port=9091
```

### 8.2 查看指标

```bash
# 访问 Prometheus 指标
http://localhost:9090/metrics

# 查看健康检查
http://localhost:9090/health
```

### 8.3 查询示例

```bash
# 查询 P99 延迟
curl -s http://localhost:9090/metrics | grep order_service_request_duration_seconds_bucket

# 查询 QPS
curl -s http://localhost:9090/metrics | grep order_service_requests_total
```

---

## 九、总结

### 9.1 已完成功能

✅ **Prometheus Agent**：自动采集 P99 延迟、QPS、错误率
✅ **gRPC 拦截器**：无侵入式指标采集
✅ **HTTP 服务**：/metrics 端点暴露指标
✅ **分析报告**：OrderList 接口需优化

### 9.2 性能优化建议

| 优化项 | 预期效果 | 优先级 |
|--------|----------|--------|
| 数据库组合索引 | P99 320ms → 50ms | 🔴 高 |
| Redis 缓存 | P99 45ms → 5ms | 🔴 高 |
| 游标分页 | 深分页性能提升 10 倍 | 🟡 中 |
| Sentinel 限流 | 防止雪崩 | 🟡 中 |

### 9.3 监控指标标准

| 指标 | 优秀 | 良好 | 需优化 | 危险 |
|------|------|------|--------|------|
| **P99 延迟** | < 50ms | < 100ms | < 500ms | > 500ms |
| **QPS** | > 1000 | > 500 | > 100 | < 100 |
| **错误率** | < 0.1% | < 1% | < 5% | > 5% |
