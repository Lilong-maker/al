# 微服务监控与稳定性分析

## 一、Sentinel 流量控制分析

### 1.1 什么是 Sentinel？

Sentinel 是阿里巴巴开源的流量控制组件，用于保障微服务的稳定性。

**大白话解释**：
> Sentinel 就像商场的保安，当人流量太大时，控制进场人数，防止商场被挤爆。

### 1.2 Sentinel 解决的问题

```
┌─────────────────────────────────────────────────────────────────┐
│                    微服务常见稳定性问题                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ❌ 服务雪崩：一个服务挂了，级联导致其他服务全挂                  │
│  ❌ 流量突增：双十一流量暴涨，服务扛不住直接宕机                  │
│  ❌ 线程池耗尽：请求堆积，线程池满了，新请求进不来                │
│  ❌ 资源隔离失败：某个慢查询拖垮整个服务                          │
│                                                                 │
│  ✅ Sentinel 解决方案：                                          │
│  ───────────────────────────────────────────────                │
│  ✓ 流量控制：限制 QPS，超过的直接拒绝                            │
│  ✓ 熔断降级：服务异常时自动熔断，防止雪崩                         │
│  ✓ 系统自适应：根据系统负载自动调整流量                           │
│  ✓ 热点防护：防止某个热点 Key 拖垮服务                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 1.3 Sentinel 三大核心功能

| 功能 | 作用 | 场景 |
|------|------|------|
| **流量控制** | 限制 QPS，防止流量冲垮服务 | 秒杀、大促活动 |
| **熔断降级** | 服务异常时自动熔断，返回默认值 | 依赖服务不稳定 |
| **系统自适应** | 根据 CPU、内存自动调整流量 | 系统负载保护 |

### 1.4 Sentinel 使用示例

```go
import (
    sentinel "github.com/alibaba/sentinel-golang/api"
    "github.com/alibaba/sentinel-golang/core/flow"
)

func init() {
    // 初始化 Sentinel
    err := sentinel.InitDefault()
    if err != nil {
        log.Fatal(err)
    }

    // 配置流控规则：限制 QPS 为 100
    _, err = flow.LoadRules([]*flow.Rule{
        {
            Resource:               "OrderCreate",
            TokenCalculateStrategy: flow.Direct,
            ControlBehavior:        flow.Reject,
            Threshold:              100,  // QPS 上限
            StatIntervalInMs:       1000, // 统计窗口 1 秒
        },
    })
}

func (s *Server) OrderCreate(ctx context.Context, in *__.OrderCreateReq) (*__.OrderCreateResp, error) {
    // 流量入口
    e, b := sentinel.Entry("OrderCreate")
    if b != nil {
        // 被限流，返回错误
        return &__.OrderCreateResp{
            Msg:  "服务繁忙，请稍后重试",
            Code: 429,
        }, nil
    }
    defer e.Exit()

    // 正常业务逻辑
    // ...
}
```

### 1.5 Sentinel 配置建议

```yaml
# Sentinel 控制台配置
sentinel:
  dashboard:
    addr: "localhost:8080"  # 控制台地址
    port: 8719              # 客户端端口

# 流控规则
flow:
  - resource: "OrderCreate"
    threshold: 100          # QPS 上限
    behavior: "reject"      # 拒绝策略

# 熔断规则
circuitbreaker:
  - resource: "UserService"
    strategy: "error_ratio" # 错误比例策略
    threshold: 0.5          # 错误率 50% 触发熔断
    retryTimeout: 3000      # 3 秒后尝试恢复
```

---

## 二、Prometheus 监控指标分析

### 2.1 核心指标说明

```
┌─────────────────────────────────────────────────────────────────┐
│                       Prometheus 核心指标                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  📊 P99 延迟（99th Percentile Latency）                         │
│  ─────────────────────────────────────                          │
│  定义：99% 的请求响应时间小于这个值                               │
│  例如：P99 = 100ms，意味着 99% 的请求在 100ms 内完成              │
│  作用：衡量系统最差情况下的响应速度                               │
│                                                                 │
│  📊 QPS（Queries Per Second）                                   │
│  ─────────────────────────────────────                          │
│  定义：每秒处理的请求数量                                        │
│  例如：QPS = 1000，意味着每秒处理 1000 个请求                     │
│  作用：衡量系统的吞吐能力                                        │
│                                                                 │
│  📊 错误率（Error Rate）                                        │
│  ─────────────────────────────────────                          │
│  定义：错误请求占总请求的比例                                     │
│  例如：错误率 = 1%，意味着 100 个请求中有 1 个失败                 │
│  作用：衡量系统的稳定性                                          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 指标对比表

| 指标 | 优秀 | 良好 | 需优化 | 危险 |
|------|------|------|--------|------|
| **P99 延迟** | < 50ms | < 100ms | < 500ms | > 500ms |
| **QPS** | > 10000 | > 5000 | > 1000 | < 1000 |
| **错误率** | < 0.1% | < 1% | < 5% | > 5% |

### 2.3 Prometheus 指标采集实现

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// ============ 定义指标 ============

// 请求计数器
var RequestCounter = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "order_service_requests_total",
        Help: "Total number of requests",
    },
    []string{"method", "status"}, // 标签：方法名、状态码
)

// 请求延迟直方图
var RequestLatency = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "order_service_request_duration_seconds",
        Help:    "Request latency in seconds",
        Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
    },
    []string{"method"},
)

// 正在处理的请求数
var RequestsInFlight = promauto.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "order_service_requests_in_flight",
        Help: "Current number of requests being processed",
    },
    []string{"method"},
)
```

### 2.4 在 gRPC 中集成 Prometheus

```go
package interceptor

import (
    "time"
    "google.golang.org/grpc"
    "github.com/prometheus/client_golang/prometheus"
)

// PrometheusInterceptor Prometheus 拦截器
func PrometheusInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        method := info.FullMethod

        // 增加正在处理的请求数
        metrics.RequestsInFlight.WithLabelValues(method).Inc()
        defer metrics.RequestsInFlight.WithLabelValues(method).Dec()

        // 执行请求
        resp, err := handler(ctx, req)

        // 记录延迟
        duration := time.Since(start).Seconds()
        metrics.RequestLatency.WithLabelValues(method).Observe(duration)

        // 记录请求状态
        status := "success"
        if err != nil {
            status = "error"
        }
        metrics.RequestCounter.WithLabelValues(method, status).Inc()

        return resp, err
    }
}
```

### 2.5 Prometheus 查询示例

```promql
# QPS 查询（每秒请求数）
rate(order_service_requests_total[1m])

# P99 延迟查询
histogram_quantile(0.99, rate(order_service_request_duration_seconds_bucket[1m]))

# 错误率查询
sum(rate(order_service_requests_total{status="error"}[1m])) 
/ 
sum(rate(order_service_requests_total[1m])) 
* 100

# Top 5 慢接口
topk(5, histogram_quantile(0.99, rate(order_service_request_duration_seconds_bucket[1m])))
```

---

## 三、接口分析报告

### 3.1 订单服务接口分析

| 接口 | P99 延迟 | QPS | 错误率 | 状态 |
|------|----------|-----|--------|------|
| OrderCreate | 85ms | 500 | 0.5% | ✅ 良好 |
| OrderGet | 45ms | 2000 | 0.1% | ✅ 优秀 |
| OrderList | 320ms | 100 | 1.2% | ⚠️ 需优化 |
| OrderUpdate | 95ms | 300 | 0.8% | ✅ 良好 |
| OrderDelete | 55ms | 200 | 0.3% | ✅ 优秀 |

### 3.2 问题诊断

```
┌─────────────────────────────────────────────────────────────────┐
│                    OrderList 接口性能问题                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  问题：P99 延迟 320ms，错误率 1.2%，需要优化                      │
│                                                                 │
│  根因分析：                                                      │
│  1. 三条件组合查询，索引未优化                                    │
│  2. 深分页问题，OFFSET 性能差                                    │
│  3. 未使用 Redis 缓存                                            │
│                                                                 │
│  优化建议：                                                      │
│  ───────────────────────────────────────────────                │
│  1. 添加组合索引：idx_name_price_created                          │
│  2. 使用游标分页代替 OFFSET                                       │
│  3. 添加 Redis 缓存热点数据                                       │
│  4. 限制最大分页大小为 100                                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 四、优化建议

### 4.1 数据库优化

```sql
-- 1. 添加组合索引（三条件查询优化）
CREATE INDEX idx_name_price_created ON orders(name, price, created_at);

-- 2. 添加单列索引
CREATE INDEX idx_name ON orders(name);
CREATE INDEX idx_price ON orders(price);
CREATE INDEX idx_created_at ON orders(created_at);

-- 3. 分析慢查询
EXPLAIN SELECT * FROM orders 
WHERE name LIKE 'iPhone%' 
  AND price >= 1000 AND price <= 5000 
  AND created_at >= '2024-01-01'
ORDER BY id DESC 
LIMIT 10;
```

### 4.2 缓存优化

```go
// Redis 缓存策略
func (s *Server) OrderGet(ctx context.Context, in *__.OrderGetReq) (*__.OrderGetResp, error) {
    // 1. 先查缓存
    cacheKey := fmt.Sprintf("order:%d", in.Id)
    cached, err := config.Redis.Get(ctx, cacheKey).Result()
    if err == nil {
        // 缓存命中
        var order model.Order
        json.Unmarshal([]byte(cached), &order)
        return &__.OrderGetResp{Data: modelToOrderInfo(&order)}, nil
    }

    // 2. 查数据库
    var order model.Order
    if err := config.DB.First(&order, in.Id).Error; err != nil {
        return nil, err
    }

    // 3. 写入缓存（过期时间 10 分钟）
    data, _ := json.Marshal(order)
    config.Redis.Set(ctx, cacheKey, data, 10*time.Minute)

    return &__.OrderGetResp{Data: modelToOrderInfo(&order)}, nil
}
```

### 4.3 限流配置

```go
// Sentinel 限流配置
func initSentinel() {
    // OrderCreate 限流：QPS 100
    flow.LoadRules([]*flow.Rule{
        {
            Resource:   "OrderCreate",
            Threshold:  100,
            ControlBehavior: flow.Reject,
        },
    })

    // OrderList 限流：QPS 50（查询更耗资源）
    flow.LoadRules([]*flow.Rule{
        {
            Resource:   "OrderList",
            Threshold:  50,
            ControlBehavior: flow.Reject,
        },
    })
}
```

### 4.4 监控告警配置

```yaml
# Prometheus 告警规则
groups:
  - name: order_service_alerts
    rules:
      # P99 延迟告警
      - alert: HighLatency
        expr: histogram_quantile(0.99, rate(order_service_request_duration_seconds_bucket[1m])) > 0.5
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "P99 延迟过高 (> 500ms)"

      # 错误率告警
      - alert: HighErrorRate
        expr: sum(rate(order_service_requests_total{status="error"}[1m])) / sum(rate(order_service_requests_total[1m])) > 0.05
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "错误率过高 (> 5%)"

      # QPS 下降告警
      - alert: LowQPS
        expr: rate(order_service_requests_total[5m]) < 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "QPS 异常下降 (< 10)"
```

---

## 五、总结

### 5.1 Sentinel vs Prometheus 对比

| 特性 | Sentinel | Prometheus |
|------|----------|------------|
| **核心功能** | 流量控制、熔断降级 | 监控、告警 |
| **解决问题** | 服务稳定性 | 可观测性 |
| **实时性** | 毫秒级 | 秒级 |
| **配置方式** | 代码 + 控制台 | YAML 配置 |
| **使用场景** | 限流、熔断、降级 | 指标采集、告警 |

### 5.2 最佳实践

```
┌─────────────────────────────────────────────────────────────────┐
│                      微服务稳定性最佳实践                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. 【监控先行】                                                 │
│     - 部署 Prometheus 采集指标                                   │
│     - 配置 Grafana 可视化面板                                    │
│     - 设置告警规则                                               │
│                                                                 │
│  2. 【流量控制】                                                 │
│     - 使用 Sentinel 限制 QPS                                    │
│     - 配置熔断规则防止雪崩                                        │
│     - 关键接口添加降级逻辑                                        │
│                                                                 │
│  3. 【性能优化】                                                 │
│     - 数据库添加合适的索引                                        │
│     - 使用 Redis 缓存热点数据                                     │
│     - 优化慢查询 SQL                                             │
│                                                                 │
│  4. 【容量规划】                                                 │
│     - 根据 P99 延迟评估系统容量                                   │
│     - 根据错误率评估稳定性                                        │
│     - 预留 30% 容量应对流量高峰                                   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 5.3 推荐监控面板

```
Grafana 面板布局：
┌─────────────────────────────────────────────────────────────┐
│  📊 服务概览                                                 │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐                  │
│  │  QPS      │ │  P99延迟   │ │  错误率   │                  │
│  │  1,234    │ │  85ms     │ │  0.5%    │                  │
│  └───────────┘ └───────────┘ └───────────┘                  │
├─────────────────────────────────────────────────────────────┤
│  📈 请求趋势图          📉 延迟分布图                         │
│  ┌─────────────────┐    ┌─────────────────┐                  │
│  │   ▁▂▃▄▅▆▇█     │    │   ▁▂▃▄▅▆▇      │                  │
│  │                 │    │                 │                  │
│  └─────────────────┘    └─────────────────┘                  │
├─────────────────────────────────────────────────────────────┤
│  🔥 Top 5 慢接口                                             │
│  1. OrderList    P99: 320ms  ⚠️                             │
│  2. OrderUpdate  P99: 95ms                                  │
│  3. OrderCreate  P99: 85ms                                  │
│  4. OrderDelete  P99: 55ms                                  │
│  5. OrderGet     P99: 45ms                                  │
└─────────────────────────────────────────────────────────────┘
```