// Package sentinel Sentinel 限流熔断模块
package sentinel

import (
	"fmt"
	"log"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/system"
)

// InitSentinel 初始化 Sentinel
func InitSentinel() error {
	// 创建默认配置
	conf := config.NewDefaultConfig()
	conf.Sentinel.Log.Dir = "./logs/sentinel"

	// 初始化 Sentinel
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		return fmt.Errorf("Sentinel 初始化失败: %v", err)
	}

	log.Println("[Sentinel] 初始化成功")
	return nil
}

// LoadFlowRules 加载流控规则（限流）
func LoadFlowRules(rules []*FlowRuleConfig) error {
	var flowRules []*flow.Rule

	for _, r := range rules {
		flowRules = append(flowRules, &flow.Rule{
			Resource:               r.Resource,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        getControlBehavior(r.ControlBehavior),
			Threshold:              r.Threshold,
			StatIntervalInMs:       r.StatIntervalInMs,
		})
	}

	_, err := flow.LoadRules(flowRules)
	if err != nil {
		return fmt.Errorf("加载流控规则失败: %v", err)
	}

	log.Printf("[Sentinel] 加载了 %d 条流控规则\n", len(flowRules))
	return nil
}

// LoadCircuitBreakerRules 加载熔断规则
func LoadCircuitBreakerRules(rules []*CircuitBreakerRuleConfig) error {
	var cbRules []*circuitbreaker.Rule

	for _, r := range rules {
		cbRules = append(cbRules, &circuitbreaker.Rule{
			Resource:         r.Resource,
			Strategy:         getCBStrategy(r.Strategy),
			RetryTimeoutMs:   r.RetryTimeoutMs,
			MinRequestAmount: r.MinRequestAmount,
			StatIntervalMs:   r.StatIntervalMs,
			MaxAllowedRtMs:   r.MaxAllowedRtMs,
			Threshold:        r.Threshold,
		})
	}

	_, err := circuitbreaker.LoadRules(cbRules)
	if err != nil {
		return fmt.Errorf("加载熔断规则失败: %v", err)
	}

	log.Printf("[Sentinel] 加载了 %d 条熔断规则\n", len(cbRules))
	return nil
}

// LoadSystemRules 加载系统规则（自适应限流）
func LoadSystemRules(rules []*SystemRuleConfig) error {
	var sysRules []*system.Rule

	for _, r := range rules {
		sysRules = append(sysRules, &system.Rule{
			MetricType:   getMetricType(r.MetricType),
			TriggerCount: r.TriggerCount,
		})
	}

	_, err := system.LoadRules(sysRules)
	if err != nil {
		return fmt.Errorf("加载系统规则失败: %v", err)
	}

	log.Printf("[Sentinel] 加载了 %d 条系统规则\n", len(sysRules))
	return nil
}

// ============ 配置结构体 ============

// FlowRuleConfig 流控规则配置
type FlowRuleConfig struct {
	Resource         string  `json:"resource"`            // 资源名
	Threshold        float64 `json:"threshold"`           // QPS 阈值
	ControlBehavior  string  `json:"control_behavior"`    // 控制行为：reject/throttling
	StatIntervalInMs uint32  `json:"stat_interval_in_ms"` // 统计窗口（毫秒）
}

// CircuitBreakerRuleConfig 熔断规则配置
type CircuitBreakerRuleConfig struct {
	Resource         string  `json:"resource"`           // 资源名
	Strategy         string  `json:"strategy"`           // 策略：slow_request/error_ratio/error_count
	RetryTimeoutMs   uint32  `json:"retry_timeout_ms"`   // 重试超时（毫秒）
	MinRequestAmount uint64  `json:"min_request_amount"` // 最小请求数
	StatIntervalMs   uint32  `json:"stat_interval_ms"`   // 统计窗口（毫秒）
	MaxAllowedRtMs   uint64  `json:"max_allowed_rt_ms"`  // 最大响应时间（慢请求策略）
	Threshold        float64 `json:"threshold"`          // 阈值
}

// SystemRuleConfig 系统规则配置
type SystemRuleConfig struct {
	MetricType   string  `json:"metric_type"`   // 指标类型：qps/thread/load/cpu
	TriggerCount float64 `json:"trigger_count"` // 触发阈值
	Strategy     string  `json:"strategy"`      // 策略
}

// ============ 辅助函数 ============

func getControlBehavior(behavior string) flow.ControlBehavior {
	switch behavior {
	case "reject":
		return flow.Reject
	case "throttling":
		return flow.Throttling
	default:
		return flow.Reject
	}
}

func getCBStrategy(strategy string) circuitbreaker.Strategy {
	switch strategy {
	case "slow_request":
		return circuitbreaker.SlowRequestRatio
	case "error_ratio":
		return circuitbreaker.ErrorRatio
	case "error_count":
		return circuitbreaker.ErrorCount
	default:
		return circuitbreaker.ErrorRatio
	}
}

func getMetricType(metricType string) system.MetricType {
	switch metricType {
	case "load":
		return system.Load
	case "cpu":
		return system.CpuUsage
	case "qps":
		return system.InboundQPS
	case "thread":
		return system.Concurrency
	default:
		return system.CpuUsage
	}
}
