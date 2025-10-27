# telegram-alerts

一个简单的 Go 服务，用来接收 TradingView 的警报 webhook 并转发到 Telegram 频道或群组。

## 配置
配置写在项目根目录的 `config.toml`（可使用 `-config` 参数指定其他路径），示例：

```toml
[telegram]
bot_token = "123456:ABCDEF"
chat_id = "-1001234567890"   # 或 "@my_channel"

[server]
addr = ":8080"               # 可选，默认为 :8080
```

字段说明：
- `telegram.bot_token`：通过 [@BotFather](https://t.me/BotFather) 创建的机器人 Token。
- `telegram.chat_id`：目标频道或群组 ID，可使用 `@channel_username` 或 `-100xxxxxxxxxx`。
- `server.addr`：HTTP 服务监听地址，留空时默认 `:8080`。

## 运行
```bash
go run ./...
# 或者指定配置文件路径
go run ./... -config /path/to/config.toml
```

服务提供两个接口：
- `POST /webhook`：TradingView 需要配置的 webhook 入口，POST JSON。
- `GET /healthz`：健康检查端点，返回 `200 OK` 和 `ok`。

## TradingView Webhook
在 TradingView 报警中，将 webhook URL 设置为你的服务地址（例如 `https://example.com/webhook`），请求体可使用：

```json
{
  "message": "Heavy volume",
  "tick": "{{ticker}}",
  "time": "{{time}}",
  "interval": "{{interval}}"
}
```

触发后，服务会发送一条包含消息正文、交易品种、周期和触发时间（若提供）的 Telegram 消息。
