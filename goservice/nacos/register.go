package nacos

import (
	"fmt"
	"log"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// 要持续发送心跳啊啊啊啊
func RegisterToNacos(serviceName, ip string, port uint64) {

	// ⭐ 关键：必须设置 BeatInterval，否则不会自动发送心跳 → 实例不健康
	clientConfig := constant.NewClientConfig(
		constant.WithNamespaceId("public"), // 默认 namespace
		constant.WithTimeoutMs(5000),
		constant.WithBeatInterval(5000), // 明确设置心跳为 5 秒
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogLevel("debug"),
	)

	// ⭐ Nacos 地址
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(
			"127.0.0.1", // Nacos IP
			8848,        // Nacos 端口
		),
	}

	// ⭐ 创建 Naming Client（带心跳 goroutine）
	client, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		log.Fatalf("Nacos client init error: %v", err)
	}

	// ⭐ 注册服务实例（必要参数必须填）
	success, err := client.RegisterInstance(vo.RegisterInstanceParam{
		ServiceName: serviceName,
		Ip:          ip,
		Port:        port,

		// ⭐ 必须：临时实例，才自动维持心跳
		Ephemeral: true,

		// ⭐ 必须：权重 > 0
		Weight: 1.0,

		Enable:  true,
		Healthy: true,

		// ⭐ 可选: 元数据
		Metadata: map[string]string{
			"version": "1.0",
			"env":     "dev",
		},
	})
	if err != nil {
		log.Fatalf("Register to Nacos failed: %v", err)
	}

	if success {
		fmt.Printf("Service %s registered to Nacos: %s:%d\n", serviceName, ip, port)
	}
}
