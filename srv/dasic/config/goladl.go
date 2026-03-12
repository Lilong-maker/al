package config

// AppConfig 全局配置结构
type AppConfig struct {
	ConSul
	Mysql        MysqlConfig  `mapstructure:"Mysql"`
	Redis        RedisConfig  `mapstructure:"Redis"`
	Nacos        NacosConfig  `mapstructure:"Nacos"`
	UserService  UserService  `mapstructure:"UserService"`
	OrderService OrderService `mapstructure:"OrderService"`
}

// MysqlConfig 数据库配置（包含连接池）
type MysqlConfig struct {
	Host            string `mapstructure:"Host"`
	Port            int    `mapstructure:"Port"`
	User            string `mapstructure:"User"`
	Password        string `mapstructure:"Password"`
	Database        string `mapstructure:"Database"`
	MaxOpenConns    int    `mapstructure:"MaxOpenConns"`
	MaxIdleConns    int    `mapstructure:"MaxIdleConns"`
	ConnMaxLifetime int    `mapstructure:"ConnMaxLifetime"`
	ConnMaxIdleTime int    `mapstructure:"ConnMaxIdleTime"`
}

// RedisConfig 缓存配置（包含连接池）
type RedisConfig struct {
	Host         string `mapstructure:"Host"`
	Port         int    `mapstructure:"Port"`
	Password     string `mapstructure:"Password"`
	Database     int    `mapstructure:"Database"`
	PoolSize     int    `mapstructure:"PoolSize"`
	MinIdleConns int    `mapstructure:"MinIdleConns"`
	MaxRetries   int    `mapstructure:"MaxRetries"`
	DialTimeout  int    `mapstructure:"DialTimeout"`
	ReadTimeout  int    `mapstructure:"ReadTimeout"`
	WriteTimeout int    `mapstructure:"WriteTimeout"`
}

// NacosConfig 配置中心配置
type NacosConfig struct {
	Addr      string `mapstructure:"Addr"`
	Port      int    `mapstructure:"Port"`
	Namespace string `mapstructure:"Namespace"`
	DataID    string `mapstructure:"DataID"`
	Group     string `mapstructure:"Group"`
	Username  string `mapstructure:"Username"`
	Password  string `mapstructure:"Password"`
}

// UserService 用户服务业务配置
type UserService struct {
	Timeout     int  `mapstructure:"Timeout"`
	MaxRetry    int  `mapstructure:"MaxRetry"`
	RateLimit   int  `mapstructure:"RateLimit"`
	EnableCache bool `mapstructure:"EnableCache"`
	CacheTTL    int  `mapstructure:"CacheTTL"`
}

// OrderService 订单服务业务配置
type OrderService struct {
	Timeout        int     `mapstructure:"Timeout"`
	MaxRetry       int     `mapstructure:"MaxRetry"`
	MaxOrderAmount float64 `mapstructure:"MaxOrderAmount"`
	EnableAsync    bool    `mapstructure:"EnableAsync"`
}

type ConSul struct {
	Host        string
	Port        int
	ServiceName string
	ServicePort int
	TTL         int
}

// ConfigObserver 配置变更观察者接口
type ConfigObserver interface {
	OnConfigChange(changedFields []string)
}

// ConfigManager 配置管理器
type ConfigManager struct {
	observers []ConfigObserver
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		observers: make([]ConfigObserver, 0),
	}
}

// RegisterObserver 注册配置变更观察者
func (cm *ConfigManager) RegisterObserver(observer ConfigObserver) {
	cm.observers = append(cm.observers, observer)
}

// NotifyObservers 通知所有观察者配置已变更
func (cm *ConfigManager) NotifyObservers(changedFields []string) {
	for _, observer := range cm.observers {
		observer.OnConfigChange(changedFields)
	}
}

// GlobalConfigManager 全局配置管理器实例
var GlobalConfigManager = NewConfigManager()

// CurrentEnv 当前运行环境（dev, beta, prod）
var CurrentEnv string

// GetDefaultMysqlConfig 获取默认 MySQL 配置
func GetDefaultMysqlConfig() MysqlConfig {
	return MysqlConfig{
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: 3600,
		ConnMaxIdleTime: 600,
	}
}

// GetDefaultRedisConfig 获取默认 Redis 配置
func GetDefaultRedisConfig() RedisConfig {
	return RedisConfig{
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,
		DialTimeout:  5,
		ReadTimeout:  3,
		WriteTimeout: 3,
	}
}
