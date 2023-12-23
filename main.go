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

// GeneralMessage 用于识别消息类型
type GeneralMessage struct {
	Type string `json:"type"`
}

// SmsMessage 用于解析短信消息
type SmsMessage struct {
	Sender      string `json:"sender"`
	SmsCode     string `json:"smsCode"`
	PhoneNumber string `json:"phoneNumber"`
	SmsMsg      string `json:"smsMsg"`
}

// CallMessage 用于解析来电消息
type CallMessage struct {
	IncomingPhoneNumber string `json:"incomingPhoneNumber"`
	PhoneNumberLocation string `json:"phoneNumberLocation"`
}

// Config 配置文件结构
type Config struct {
	Broker              string `yaml:"broker"`
	TopicSms            string `yaml:"topicSms"`
	TopicCall           string `yaml:"topicCall"`
	WxAPI               string `yaml:"wxAPI"`
	RecipientWxNickname string `yaml:"recipientWxNickname"`
}

var cfg Config
var logger *log.Logger

func init() {
	// 初始化日志，输出到控制台
	logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

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

// 处理消息的函数
var messagePubHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	// 日志记录收到的消息主题
	logger.Printf("收到新消息: %s, 消息内容: %s", msg.Topic(), msg.Payload())

	// 尝试解析为通用消息格式以获取消息类型
	var generalMsg GeneralMessage
	err := json.Unmarshal(msg.Payload(), &generalMsg)
	if err != nil {
		logger.Printf("消息解析错误: %v\n", err)
		return
	}

	// 根据消息类型分别处理
	switch generalMsg.Type {
	case "sms":
		var smsMsg SmsMessage
		err := json.Unmarshal(msg.Payload(), &smsMsg)
		if err != nil {
			logger.Printf("短信消息解析错误: %v\n", err)
			return
		}
		handleSmsMessage(smsMsg)

	case "call":
		var callMsg CallMessage
		err := json.Unmarshal(msg.Payload(), &callMsg)
		if err != nil {
			logger.Printf("来电消息解析错误: %v\n", err)
			return
		}
		handleCallMessage(callMsg)

	default:
		logger.Printf("未知的消息类型: %s", generalMsg.Type)
	}
}

// 处理短信消息的函数
func handleSmsMessage(smsMsg SmsMessage) {
	var content string
	if smsMsg.SmsCode != "" {
		content = fmt.Sprintf("来自: %s, 验证码: %s, 发送者: %s", smsMsg.PhoneNumber, smsMsg.SmsCode, smsMsg.Sender)
	} else {
		content = fmt.Sprintf("来自: %s, 内容: %s", smsMsg.PhoneNumber, smsMsg.SmsMsg)
	}
	// 发送消息给微信
	wxMessagePub(content)
}

// 处理来电消息的函数
func handleCallMessage(callMsg CallMessage) {
	content := fmt.Sprintf("来电号码: %s, 归宿地: %s", callMsg.IncomingPhoneNumber, callMsg.PhoneNumberLocation)
	// 发送消息给微信
	wxMessagePub(content)
}

// 发送消息到微信的函数
var wxMessagePub = func(content string) {
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

	// 处理响应
	logger.Printf("请求发送成功, 响应状态码: %d\n", resp.StatusCode)
}

// MQTT连接成功的处理函数
var connectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
	logger.Println("已连接到MQTT服务器")
}

// MQTT连接丢失的处理函数
var connectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
	logger.Printf("MQTT连接丢失: %v", err)
}

// 生成随机客户端ID的函数
func generateRandomClientID(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		logger.Fatalf("生成随机客户端ID失败: %v", err)
	}
	return hex.EncodeToString(bytes)
}

func main() {
	// MQTT客户端设置
	opts := MQTT.NewClientOptions().AddBroker(cfg.Broker)
	randomClientID := generateRandomClientID(10) // 生成10字节长度的随机ID
	opts.SetClientID(randomClientID)
	opts.SetDefaultPublishHandler(messagePubHandler)

	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	// 创建并连接MQTT客户端
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// 订阅短信和来电消息的主题
	subscribe(client, cfg.TopicSms)
	subscribe(client, cfg.TopicCall)

	// 阻塞主线程，防止程序退出
	select {}
}

// 订阅主题的函数
func subscribe(client MQTT.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	logger.Printf("已订阅主题: %s\n", topic)
}
