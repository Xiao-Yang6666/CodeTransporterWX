# 一、项目介绍

CodeTransporter是一款专注于解决电脑端登录网站或应用时验证码接收不便的安卓工具应用。该项目通过监听系统短信广播，获取手机接收到的验证码消息后，将其推送到MQTT服务器。Windows端和微信端通过订阅MQTT服务器上的相关主题，实现了验证码的即时推送，用户可在任何时候通过订阅端查看验证码，省去了繁琐的手机操作。

## 主要特点与优势

1. **验证码监听与推送：** CodeTransporter通过系统短信广播监听手机接收到的验证码消息，并将其即时推送到MQTT服务器。Windows端和微信端通过订阅MQTT服务器上的主题，实现了验证码的方便查看，用户无需频繁查看手机，省去了繁琐的操作。
2. **后台运行无干扰：** 应用无需一直挂在后台，通过高效的短信广播接收方式，确保验证码的实时推送。这样既减少了对手机性能的额外消耗，又保障了用户的使用体验。
3. **持久可靠：** CodeTransporter的设计保证了应用的持久性，即使手机重启，用户无需二次操作或手动开启应用，仍能正常接收并推送验证码至MQTT服务器。
4. **解决手机不在身边问题：** 用户只需在Windows平台或微信端上订阅MQTT服务器的主题，即可在任何地方获取到实时的验证码信息，解决了因手机不在身边而无法及时接收验证码的问题。

## 使用场景

- 在电脑端登录各类网站或应用，无需频繁查看手机，提高了操作的便捷性。
- 解决手机不在身边，但仍需获取验证码的场景，如在办公室、家中等。

CodeTransporter是一款简单而实用的工具，通过MQTT服务器的推送机制，为用户提供了验证码接收的便捷解决方案。用户可通过订阅端实时查看验证码，提升了整体使用体验。



# 二、数据流程

![image-20231223153548558](C:\Users\xiaoyang\Documents\学习\md文档\0.image\image-20231223153548558.png)



# 三、部署步骤

## 1、emqx部署

EMQX 是一款[开源](https://github.com/emqx/emqx)的大规模分布式 MQTT 消息服务器，功能丰富，专为物联网和实时通信应用而设计。EMQX 5.0 单集群支持 MQTT 并发连接数高达 1 亿条，单服务器的传输与处理吞吐量可达每秒百万级 MQTT 消息，同时保证毫秒级的低时延。

```shell
docker run -d --name emqx -p 1883:1883 -p 8083:8083 -p 8084:8084 -p 8883:8883 -p 18083:18083 emqx/emqx:latest
```

项目地址：https://www.emqx.io/docs/zh/latest/

通过浏览器访问 http://localhost:18083/（localhost 可替换为您的实际 IP 地址）以访问 [EMQX Dashboard](https://www.emqx.io/docs/zh/latest/dashboard/introduction.html) 管理控制台，进行设备连接与相关指标监控管理。

 默认用户名及密码：admin、public



## 2、wechatbot部署

仓库地址：https://github.com/danni-cool/docker-wechatbot-webhook

```shell
# 拉取镜像
sudo docker pull dannicool/docker-wechatbot-webhook
# 自定义一下apiToken，后期调用需要使用
sudo docker run -d \
--name wxBotWebhook \
-p 3001:3001 \
-e LOGIN_API_TOKEN="xiaoyang" \
dannicool/docker-wechatbot-webhook
```

登录，在浏览器打开下面连接进行扫码登录。（localhost换成你的服务器ip）

https://localhost:3001/login?token=YOUR_PERSONAL_TOKEN



## 3、CodeTransporter部署

### 1、安卓端部署

仓库地址：https://github.com/Xiao-Yang6666/CodeTransporterAndroid.git

1. 手机端安装软件，手动赋予读取短信权限，自启动权限。

2. 访问http://192.168.2.251:8848/code/index.html，通过url的方式打开软件的设置页面，进行相关设置。

   <img src="C:\Users\xiaoyang\AppData\Roaming\Typora\typora-user-images\image-20231223154341178.png" alt="image-20231223154341178" style="zoom: 25%;" />

### 2、win端部署

仓库地址：https://github.com/Xiao-Yang6666/CodeTransporterWin.git

#### 2.1、yaml文件配置

```yaml
# mqtt服务器地址
broker: "tcp://124.222.11.237:1883"
# mqtt订阅队列
topic: "sms/topic"
```

#### 2.2、直接运行exe

> 注: yaml配置文件需要跟可执行文件，在同一目录下。



### 3、wx机器人端部署

仓库地址：https://github.com/Xiao-Yang6666/CodeTransporterWX.git

主要依赖于wechatbot项目：https://github.com/danni-cool/docker-wechatbot-webhook

#### 3.1、yaml文件配置

```yaml
# mqtt服务器地址
broker: "tcp://124.222.11.237:1883"
# mqtt订阅队列
topic: "sms/topic"
# wechatbot部署的接口地址
wxAPI: "http://192.168.2.100:3001"
# 接收者的微信昵称
recipientWxNickname: "晓阳"
```

#### 3.2、直接后台运行

```shell
nohup ./codeMqttWX &
```

>  注: yaml配置文件需要跟可执行文件，在同一目录下。

