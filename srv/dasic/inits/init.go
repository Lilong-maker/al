package inits

import (
	"al/srv/dasic/config"
	"al/srv/handler/model"
	"al/srv/handler/service"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	// 第一步：读取本地配置（Nacos 连接信息）
	ViperInit()

	// 第二步：根据环境变量设置 Group
	setEnvironment()

	// 第三步：从 Nacos 加载对应环境的配置
	if err := NacosInit(); err != nil {
		fmt.Printf("Nacos 配置加载失败: %v\n", err)
	} else {
		fmt.Printf("Nacos 配置加载成功 (Group: %s)\n", config.Gen.Nacos.Group)
	}

	// 第四步：初始化 MySQL
	MysqlInit()

	// 第五步：初始化 Redis
	RedisInit()

	// 第六步：初始化热配置演示服务
	service.InitHotConfigDemo()
}

var err error

// setEnvironment 根据环境变量设置配置分组
func setEnvironment() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	config.CurrentEnv = env

	switch env {
	case "dev":
		config.Gen.Nacos.Group = "dev"
	case "beta":
		config.Gen.Nacos.Group = "beta"
	case "prod":
		config.Gen.Nacos.Group = "prod"
	default:
		config.Gen.Nacos.Group = "dev"
	}

	dataID := os.Getenv("APP_DATA_ID")
	if dataID != "" {
		config.Gen.Nacos.DataID = dataID
	}

	fmt.Printf("========================================\n")
	fmt.Printf("环境配置:\n")
	fmt.Printf("  - 当前环境: %s\n", env)
	fmt.Printf("  - Nacos Group: %s\n", config.Gen.Nacos.Group)
	fmt.Printf("  - Nacos DataID: %s\n", config.Gen.Nacos.DataID)
	fmt.Printf("========================================\n")
}

// MysqlInit 初始化 MySQL 连接（包含连接池配置）
func MysqlInit() {
	mysqlConfig := config.Gen.Mysql

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlConfig.User,
		mysqlConfig.Password,
		mysqlConfig.Host,
		mysqlConfig.Port,
		mysqlConfig.Database,
	)

	config.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}

	sqlDB, err := config.DB.DB()
	if err != nil {
		panic(fmt.Sprintf("获取数据库连接失败: %v", err))
	}

	maxOpenConns := mysqlConfig.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 100
	}

	maxIdleConns := mysqlConfig.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 10
	}

	connMaxLifetime := mysqlConfig.ConnMaxLifetime
	if connMaxLifetime == 0 {
		connMaxLifetime = 3600
	}

	connMaxIdleTime := mysqlConfig.ConnMaxIdleTime
	if connMaxIdleTime == 0 {
		connMaxIdleTime = 600
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(connMaxIdleTime) * time.Second)

	fmt.Println("MySQL 连接成功")
	fmt.Printf("MySQL 连接池: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%ds\n",
		maxOpenConns, maxIdleConns, connMaxLifetime)

	err = config.DB.AutoMigrate(&model.Order{})
	if err != nil {
		fmt.Printf("表迁移失败: %v\n", err)
		return
	}
	fmt.Println("表迁移成功")
}

// RedisInit 初始化 Redis 连接（包含连接池配置）
func RedisInit() {
	redisConfig := config.Gen.Redis

	// 创建 Redis 客户端
	config.Redis = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password:     redisConfig.Password,
		DB:           redisConfig.Database,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
		MaxRetries:   redisConfig.MaxRetries,
		DialTimeout:  time.Duration(redisConfig.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(redisConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(redisConfig.WriteTimeout) * time.Second,
	})

	// 如果配置为 0，使用默认值
	if config.Redis.Options().PoolSize == 0 {
		config.Redis.Options().PoolSize = 100
	}
	if config.Redis.Options().MinIdleConns == 0 {
		config.Redis.Options().MinIdleConns = 10
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = config.Redis.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Redis 连接失败: %v\n", err)
		return
	}

	fmt.Println("Redis 连接成功")
	fmt.Printf("Redis 连接池: PoolSize=%d, MinIdle=%d, Addr=%s:%d\n",
		redisConfig.PoolSize, redisConfig.MinIdleConns, redisConfig.Host, redisConfig.Port)
}

// ViperInit 初始化本地配置文件（只读取 Nacos 连接信息）
func ViperInit() {
	viper.SetConfigFile("C:\\Users\\Lenovo\\Desktop\\al\\config.yml")
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("读取本地配置文件失败: %v\n", err)
		return
	}
	err = viper.Unmarshal(&config.Gen)
	if err != nil {
		fmt.Printf("解析本地配置失败: %v\n", err)
		return
	}
	fmt.Println("本地配置文件加载成功")
}
