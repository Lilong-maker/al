package init

import (
	"al/srv/dasic/config"
	"fmt"
	"reflect"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
)

var nacosClient config_client.IConfigClient

// lastConfig 用于保存上一次配置，用于对比变更
var lastConfig config.AppConfig

// NacosInit 初始化 Nacos 配置中心连接
func NacosInit() error {
	nacosConfig := config.Gen.Nacos

	// 检查配置是否有效
	if nacosConfig.Addr == "" {
		return fmt.Errorf("Nacos 地址未配置")
	}

	// 创建 serverConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: nacosConfig.Addr,
			Port:   uint64(nacosConfig.Port),
		},
	}

	// 创建 clientConfig
	clientConfig := constant.ClientConfig{
		NamespaceId:         nacosConfig.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "./nacos/log",
		CacheDir:            "./nacos/cache",
		LogLevel:            "info",
		Username:            nacosConfig.Username,
		Password:            nacosConfig.Password,
	}

	// 创建配置客户端
	var err error
	nacosClient, err = clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return fmt.Errorf("创建 Nacos 客户端失败: %v", err)
	}

	// 从 Nacos 获取配置
	content, err := nacosClient.GetConfig(vo.ConfigParam{
		DataId: nacosConfig.DataID,
		Group:  nacosConfig.Group,
	})
	if err != nil {
		return fmt.Errorf("从 Nacos 获取配置失败: %v", err)
	}

	// 检查配置是否为空
	if content == "" || content == "config data not exist" {
		return fmt.Errorf("Nacos 配置为空，请先在控制台创建配置")
	}

	fmt.Printf("从 Nacos 获取配置成功，内容长度: %d\n", len(content))

	// 使用 viper 解析配置
	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(strings.NewReader(content)); err != nil {
		return fmt.Errorf("解析 Nacos 配置失败: %v", err)
	}

	// 更新配置到全局变量
	if err := viper.Unmarshal(&config.Gen); err != nil {
		return fmt.Errorf("更新配置失败: %v", err)
	}

	// 保存初始配置用于后续对比
	lastConfig = *config.Gen

	fmt.Println("Nacos 配置加载成功")
	fmt.Printf("当前 UserService 配置: Timeout=%d, MaxRetry=%d, RateLimit=%d, EnableCache=%v\n",
		config.Gen.UserService.Timeout,
		config.Gen.UserService.MaxRetry,
		config.Gen.UserService.RateLimit,
		config.Gen.UserService.EnableCache)

	// 监听配置变化
	err = nacosClient.ListenConfig(vo.ConfigParam{
		DataId: nacosConfig.DataID,
		Group:  nacosConfig.Group,
		OnChange: func(namespace, group, dataID, data string) {
			fmt.Println("========================================")
			fmt.Println("检测到 Nacos 配置变化，正在重新加载...")
			fmt.Println("========================================")

			// 重新解析配置
			viper.SetConfigType("yaml")
			if err := viper.ReadConfig(strings.NewReader(data)); err != nil {
				fmt.Printf("重新解析配置失败: %v\n", err)
				return
			}

			// 临时保存新配置
			newConfig := config.AppConfig{}
			if err := viper.Unmarshal(&newConfig); err != nil {
				fmt.Printf("解析新配置失败: %v\n", err)
				return
			}

			// 检测变更的字段
			changedFields := detectConfigChanges(&lastConfig, &newConfig)

			// 更新全局配置
			*config.Gen = newConfig

			// 通知所有观察者
			if len(changedFields) > 0 {
				fmt.Printf("配置变更字段: %v\n", changedFields)
				config.GlobalConfigManager.NotifyObservers(changedFields)
			}

			// 更新 lastConfig 用于下一次对比
			lastConfig = newConfig

			fmt.Println("配置重新加载成功")
			fmt.Printf("更新后 UserService 配置: Timeout=%d, MaxRetry=%d, RateLimit=%d, EnableCache=%v\n",
				config.Gen.UserService.Timeout,
				config.Gen.UserService.MaxRetry,
				config.Gen.UserService.RateLimit,
				config.Gen.UserService.EnableCache)
		},
	})

	if err != nil {
		return fmt.Errorf("监听 Nacos 配置失败: %v", err)
	}

	fmt.Println("Nacos 配置监听已启动")
	return nil
}

// detectConfigChanges 检测配置变更的字段
// 返回变更的字段路径列表，如 ["UserService.Timeout", "OrderService.MaxRetry"]
func detectConfigChanges(oldConfig, newConfig *config.AppConfig) []string {
	var changedFields []string

	// 对比 UserService
	changedFields = append(changedFields, compareStruct(
		"UserService",
		reflect.ValueOf(&oldConfig.UserService),
		reflect.ValueOf(&newConfig.UserService),
	)...)

	// 对比 OrderService
	changedFields = append(changedFields, compareStruct(
		"OrderService",
		reflect.ValueOf(&oldConfig.OrderService),
		reflect.ValueOf(&newConfig.OrderService),
	)...)

	return changedFields
}

// compareStruct 对比两个结构体的字段差异
func compareStruct(prefix string, oldVal, newVal reflect.Value) []string {
	var changes []string

	// 解引用指针
	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
		newVal = newVal.Elem()
	}

	// 获取结构体类型
	oldType := oldVal.Type()

	for i := 0; i < oldVal.NumField(); i++ {
		field := oldType.Field(i)
		oldFieldVal := oldVal.Field(i)
		newFieldVal := newVal.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		// 对比字段值
		if !reflect.DeepEqual(oldFieldVal.Interface(), newFieldVal.Interface()) {
			fieldPath := prefix + "." + field.Name
			changes = append(changes, fieldPath)
			fmt.Printf("  [变更] %s: %v -> %v\n",
				fieldPath,
				oldFieldVal.Interface(),
				newFieldVal.Interface())
		}
	}

	return changes
}
