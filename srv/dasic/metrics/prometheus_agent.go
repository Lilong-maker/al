// Package metrics Prometheus 指标采集模块
package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusAgent Prometheus 指标采集器
type PrometheusAgent struct {
	// 请求计数器（按方法和状态码分类）
	RequestCounter *prometheus.CounterVec

	// 请求延迟直方图（用于计算 P99/P95/P50）
	RequestLatency *prometheus.HistogramVec

	// 正在处理的请求数（并发数）
	RequestsInFlight *prometheus.GaugeVec

	// 错误计数器
	ErrorCounter *prometheus.CounterVec

	// 响应大小（字节）
	ResponseSize *prometheus.HistogramVec

	// 请求大小（字节）
	RequestSize *prometheus.HistogramVec
}

// GlobalPrometheusAgent 全局 Prometheus Agent 实例
var GlobalPrometheusAgent *PrometheusAgent

// InitPrometheus 初始化 Prometheus 指标
func InitPrometheus() {
	GlobalPrometheusAgent = &PrometheusAgent{
		// 请求计数器
		RequestCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "order_service_requests_total",
				Help: "Total number of requests",
			},
			[]string{"method", "status"},
		),

		// 请求延迟直方图（桶分布用于计算 P99）
		RequestLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "order_service_request_duration_seconds",
				Help:    "Request latency in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
			},
			[]string{"method"},
		),

		// 正在处理的请求数
		RequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "order_service_requests_in_flight",
				Help: "Current number of requests being processed",
			},
			[]string{"method"},
		),

		// 错误计数器（按错误类型分类）
		ErrorCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "order_service_errors_total",
				Help: "Total number of errors",
			},
			[]string{"method", "error_type"},
		),

		// 响应大小
		ResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "order_service_response_size_bytes",
				Help:    "Response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method"},
		),

		// 请求大小
		RequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "order_service_request_size_bytes",
				Help:    "Request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method"},
		),
	}

	fmt.Println("[Prometheus] 指标采集器初始化成功")
}

// RecordRequest 记录请求指标
func (pa *PrometheusAgent) RecordRequest(method string, duration time.Duration, status string) {
	pa.RequestCounter.WithLabelValues(method, status).Inc()
	pa.RequestLatency.WithLabelValues(method).Observe(duration.Seconds())
}

// RecordError 记录错误指标
func (pa *PrometheusAgent) RecordError(method string, errorType string) {
	pa.ErrorCounter.WithLabelValues(method, errorType).Inc()
}

// IncInFlight 增加正在处理的请求数
func (pa *PrometheusAgent) IncInFlight(method string) {
	pa.RequestsInFlight.WithLabelValues(method).Inc()
}

// DecInFlight 减少正在处理的请求数
func (pa *PrometheusAgent) DecInFlight(method string) {
	pa.RequestsInFlight.WithLabelValues(method).Dec()
}

// GetMetricsSummary 获取指标摘要（用于分析）
func (pa *PrometheusAgent) GetMetricsSummary() map[string]interface{} {
	return map[string]interface{}{
		"status": "running",
		"time":   time.Now().Format("2024-01-01 12:00:00"),
	}
}

// AnalysisResult 分析结果结构
type AnalysisResult struct {
	Method      string  `json:"method"`
	P99Latency  float64 `json:"p99_latency"` // P99 延迟（毫秒）
	P95Latency  float64 `json:"p95_latency"` // P95 延迟（毫秒）
	P50Latency  float64 `json:"p50_latency"` // P50 延迟（毫秒）
	QPS         float64 `json:"qps"`         // QPS
	ErrorRate   float64 `json:"error_rate"`  // 错误率（%）
	TotalReq    int64   `json:"total_req"`   // 总请求数
	TotalErr    int64   `json:"total_err"`   // 总错误数
	AvgLatency  float64 `json:"avg_latency"` // 平均延迟（毫秒）
	Status      string  `json:"status"`      // 状态：优秀/良好/需优化/危险
	Suggestions string  `json:"suggestions"` // 优化建议
}

// AnalyzeMetrics 分析指标并生成报告
func AnalyzeMetrics(method string) *AnalysisResult {
	// 这里可以从 Prometheus 查询真实数据
	// 目前返回示例数据
	result := &AnalysisResult{
		Method:     method,
		P99Latency: 85.0,
		P95Latency: 60.0,
		P50Latency: 25.0,
		QPS:        500.0,
		ErrorRate:  0.5,
		TotalReq:   10000,
		TotalErr:   50,
		AvgLatency: 45.0,
	}

	// 根据指标判断状态
	result.Status = judgeStatus(result.P99Latency, result.QPS, result.ErrorRate)
	result.Suggestions = generateSuggestions(result)

	return result
}

// judgeStatus 根据指标判断状态
func judgeStatus(p99 float64, qps float64, errorRate float64) string {
	if p99 < 50 && qps > 1000 && errorRate < 0.1 {
		return "优秀 🌟"
	} else if p99 < 100 && qps > 500 && errorRate < 1.0 {
		return "良好 👍"
	} else if p99 < 500 && qps > 100 && errorRate < 5.0 {
		return "需优化 ⚠️"
	}
	return "危险 🔥"
}

// generateSuggestions 生成优化建议
func generateSuggestions(result *AnalysisResult) string {
	var suggestions []string

	if result.P99Latency > 100 {
		suggestions = append(suggestions, "P99延迟过高，建议：1)添加数据库索引 2)使用Redis缓存 3)优化SQL查询")
	}

	if result.QPS < 100 {
		suggestions = append(suggestions, "QPS较低，建议：1)检查是否有性能瓶颈 2)增加缓存 3)优化数据库连接池")
	}

	if result.ErrorRate > 1.0 {
		suggestions = append(suggestions, "错误率较高，建议：1)添加错误日志 2)增加重试机制 3)检查依赖服务状态")
	}

	if len(suggestions) == 0 {
		return "当前性能良好，无需优化"
	}

	return fmt.Sprintf("发现 %d 个问题：%s", len(suggestions), suggestions)
}

// PrintAnalysisReport 打印分析报告
func PrintAnalysisReport(results []*AnalysisResult) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                  Prometheus 接口分析报告                     ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Println("║  接口          P99延迟   QPS      错误率    状态              ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	for _, r := range results {
		fmt.Printf("║  %-12s  %6.1fms  %6.1f   %5.2f%%   %-10s ║\n",
			r.Method, r.P99Latency, r.QPS, r.ErrorRate, r.Status)
	}

	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// MonitorContext 监控上下文
type MonitorContext struct {
	Method    string
	StartTime time.Time
}

// StartMonitor 开始监控
func StartMonitor(method string) *MonitorContext {
	if GlobalPrometheusAgent != nil {
		GlobalPrometheusAgent.IncInFlight(method)
	}
	return &MonitorContext{
		Method:    method,
		StartTime: time.Now(),
	}
}

// FinishMonitor 结束监控
func (mc *MonitorContext) FinishMonitor(status string) time.Duration {
	duration := time.Since(mc.StartTime)

	if GlobalPrometheusAgent != nil {
		GlobalPrometheusAgent.DecInFlight(mc.Method)
		GlobalPrometheusAgent.RecordRequest(mc.Method, duration, status)
	}

	return duration
}

// RecordErrorWithContext 记录错误
func RecordErrorWithContext(method string, errorType string) {
	if GlobalPrometheusAgent != nil {
		GlobalPrometheusAgent.RecordError(method, errorType)
	}
}
