package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// SmsMessage 定义一个结构体来映射JSON数据
type SmsMessage struct {
	Sender      string `json:"sender"`
	SmsCode     string `json:"smsCode"`
	PhoneNumber string `json:"phoneNumber"`
	SmsMsg      string `json:"smsMsg"`
}

// Config 定义配置文件结构
type Config struct {
	Broker              string `yaml:"broker"`
	Topic               string `yaml:"topic"`
	WxAPI               string `yaml:"wxAPI"`
	RecipientWxNickname string `yaml:"recipientWxNickname"`
}

var cfg Config
var logger *log.Logger

func init() {
	// 初始化日志
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败:", err)
	}
	logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	// 读取配置文件
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		logger.Fatalf("读取配置文件失败: %v", err)
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		logger.Fatalf("解析配置文件失败: %v", err)
	}
}

// 当收到消息时的处理函数
var messagePubHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	var smsMsg SmsMessage
	err := json.Unmarshal(msg.Payload(), &smsMsg)
	if err != nil {
		logger.Printf("消息解析错误: %v\n", err)
		return
	}

	var content string
	if smsMsg.SmsCode != "" {
		content = fmt.Sprintf("来自: %s, 验证码: %s, 发送者: %s", smsMsg.PhoneNumber, smsMsg.SmsCode, smsMsg.Sender)
	} else {
		content = fmt.Sprintf("来自: %s, 内容: %s", smsMsg.PhoneNumber, smsMsg.SmsMsg)
	}

	// 构造要发送的数据
	data := map[string]string{
		"to":      cfg.RecipientWxNickname,
		"type":    "text",
		"content": content,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Printf("构造请求数据失败: %v\n", err)
		return
	}

	// 发送HTTP POST请求
	resp, err := http.Post(cfg.WxAPI+"/webhook/msg", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Printf("发送HTTP请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 可以在此处处理响应，例如打印状态码
	logger.Printf("请求发送成功, 响应状态码: %d\n", resp.StatusCode)
}

var connectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
	logger.Println("已连接")
}

var connectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
	logger.Printf("连接丢失: %v", err)
}

func generateRandomClientID(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// 处理错误
		logger.Fatalf("生成随机客户端ID失败: %v", err)
	}
	return hex.EncodeToString(bytes)
}

func main() {
	// MQTT订阅
	opts := MQTT.NewClientOptions().AddBroker(cfg.Broker)
	randomClientID := generateRandomClientID(10) // 生成10字节长度的随机ID
	opts.SetClientID(randomClientID)
	opts.SetDefaultPublishHandler(messagePubHandler)

	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	subscribe(client, cfg.Topic)

	select {}
}

func subscribe(client MQTT.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	logger.Printf("已订阅主题: %s\n", topic)
}
