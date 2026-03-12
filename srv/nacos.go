package main

//
//import (
//	"fmt"
//	"srv/dasic/config"
//	"strings"
//
//	"github.com/nacos-group/nacos-sdk-go/v2/clients"
//	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
//	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
//	"github.com/nacos-group/nacos-sdk-go/v2/vo"
//	"github.com/spf13/viper"
//)
//
//var nacosClient config_client.IConfigClient
//
//// NacosInit 初始化 Nacos 配置中心连接
//func NacosInit() error {
//	nacosConfig := config.Gen.Nacos
//
//	// 检查配置是否有效
//	if nacosConfig.Addr == "" {
//		return fmt.Errorf("Nacos 地址未配置")
//	}
//
//	// 创建 serverConfig
//	serverConfigs := []constant.ServerConfig{
//		{
//			IpAddr: nacosConfig.Addr,
//			Port:   uint64(nacosConfig.Port),
//		},
//	}
//
//	// 创建 clientConfig
//	clientConfig := constant.ClientConfig{
//		NamespaceId:         nacosConfig.Namespace,
//		TimeoutMs:           5000,
//		NotLoadCacheAtStart: true,
//		LogDir:              "./nacos/log",
//		CacheDir:            "./nacos/cache",
//		LogLevel:            "info",
//		Username:            nacosConfig.Username,
//		Password:            nacosConfig.Password,
//	}
//
//	// 创建配置客户端
//	var err error
//	nacosClient, err = clients.CreateConfigClient(map[string]interface{}{
//		"serverConfigs": serverConfigs,
//		"clientConfig":  clientConfig,
//	})
//	if err != nil {
//		return fmt.Errorf("创建 Nacos 客户端失败: %v", err)
//	}
//
//	// 从 Nacos 获取配置
//	content, err := nacosClient.GetConfig(vo.ConfigParam{
//		DataId: nacosConfig.DataID,
//		Group:  nacosConfig.Group,
//	})
//	if err != nil {
//		return fmt.Errorf("从 Nacos 获取配置失败: %v", err)
//	}
//
//	// 检查配置是否为空
//	if content == "" || content == "config data not exist" {
//		return fmt.Errorf("Nacos 配置为空，请先在控制台创建配置")
//	}
//
//	fmt.Printf("从 Nacos 获取配置成功，内容长度: %d\n", len(content))
//
//	// 使用 viper 解析配置
//	viper.SetConfigType("yaml")
//	if err := viper.ReadConfig(strings.NewReader(content)); err != nil {
//		return fmt.Errorf("解析 Nacos 配置失败: %v", err)
//	}
//
//	// 更新配置到全局变量
//	if err := viper.Unmarshal(&config.Gen); err != nil {
//		return fmt.Errorf("更新配置失败: %v", err)
//	}
//
//	fmt.Println("Nacos 配置加载成功")
//
//	// 监听配置变化
//	err = nacosClient.ListenConfig(vo.ConfigParam{
//		DataId: nacosConfig.DataID,
//		Group:  nacosConfig.Group,
//		OnChange: func(namespace, group, dataID, data string) {
//			fmt.Println("检测到 Nacos 配置变化，正在重新加载...")
//
//			// 重新解析配置
//			viper.SetConfigType("yaml")
//			if err := viper.ReadConfig(strings.NewReader(data)); err != nil {
//				fmt.Printf("重新解析配置失败: %v\n", err)
//				return
//			}
//
//			// 更新配置
//			if err := viper.Unmarshal(&config.Gen); err != nil {
//				fmt.Printf("更新配置失败: %v\n", err)
//				return
//			}
//
//			fmt.Println("配置重新加载成功")
//		},
//	})
//
//	if err != nil {
//		return fmt.Errorf("监听 Nacos 配置失败: %v", err)
//	}
//
//	fmt.Println("Nacos 配置监听已启动")
//	return nil
//}
