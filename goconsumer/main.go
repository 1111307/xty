package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
	// 1. 初始化 Nacos 客户端配置
	clientConfig := constant.NewClientConfig()
	serverConfigs := []constant.ServerConfig{
		// Nacos Server 地址必须与您的配置一致
		*constant.NewServerConfig("127.0.0.1", 8848),
	}

	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		log.Fatalf("Nacos client init error: %v", err)
	}

	serviceName := "goservice"
	endpoint := "/api/echo"

	fmt.Printf("--- 步骤一：通过 Nacos 发现服务 [%s] ---\n", serviceName)

	// 2. 从 Nacos 中选择一个健康的服务实例
	// 强制指定 GroupName 为 DEFAULT_GROUP，以避免查找错误
	instance, err := namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   "DEFAULT_GROUP", // 确保与 goservice 注册时使用的组名一致
	})
	if err != nil {
		log.Fatalf("Failed to select healthy instance from Nacos: healthy instance list is empty! Error: %v", err)
	}

	// 3. 构建完整的服务 URL
	serviceUrl := fmt.Sprintf("http://%s:%d%s", instance.Ip, instance.Port, endpoint)

	fmt.Printf("✅ Nacos 发现实例成功: IP=%s, Port=%d\n", instance.Ip, instance.Port)
	fmt.Printf("✅ 完整调用 URL: %s\n", serviceUrl)

	fmt.Println("--- 步骤二：向发现的实例发起 HTTP POST 请求 ---")

	// 4. 准备请求体 (Body)
	requestBody := map[string]interface{}{
		"consumer": "Nacos Go Client",
		"time":     time.Now().Format(time.RFC3339),
	}
	jsonBody, _ := json.Marshal(requestBody)

	// 5. 发送 POST 请求
	// 设置 5 秒超时，避免请求长时间阻塞
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(serviceUrl, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Fatalf("HTTP request failed (Provider may be DOWN): %v", err)
	}
	defer resp.Body.Close()

	// 6. 解析并打印响应
	var result map[string]interface{}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("HTTP call failed. Status: %s. Check if /api/echo handler is correct.", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	fmt.Printf("--- 成功接收到服务响应 (Status: %s) ---\n", resp.Status)
	// 格式化打印 JSON
	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(prettyJSON))
}
