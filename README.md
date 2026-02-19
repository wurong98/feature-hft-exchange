# HFT 模拟交易所

基于币安真实交易数据流的永续合约模拟交易所，支持 CCXT 兼容 API 和策略排行榜。

## 核心特性

- **真实市场数据**: 实时订阅币安 WebSocket 交易流，撮合基于真实价格
- **永续合约模拟**: 支持多空双向交易、杠杆开仓、持仓管理
- **CCXT 兼容 API**: 使用标准 Binance API 格式，可直接用 CCXT 库连接
- **Web Dashboard**: 实时策略排行榜，展示各策略收益情况
- **零成本模拟**: 无需真实资金，通过 API Key 隔离各策略账户

## 技术栈

- **后端**: Go 1.21+ / Gin / SQLite
- **数据源**: 币安 WebSocket (wss://stream.binance.com:9443/ws)
- **前端**: 原生 HTML/JS (Vue.js 可选升级)
- **API 协议**: REST + WebSocket (CCXT 兼容)

## 快速开始

### 1. 构建

```bash
# 克隆项目
git clone <repository-url>
cd ShadowHFT

# 构建主程序
go build -o bin/hft-sim .

# 构建管理工具
go build -o bin/admin ./cmd/admin
```

### 2. 创建 API Key

```bash
# 创建策略账户（初始资金 10000 USDT）
./bin/admin -action=create -name="MyStrategy" -desc="测试策略" -balance=10000

# 列出所有 API Key
./bin/admin -action=list

# 删除 API Key
./bin/admin -action=delete -key="<your-api-key>"
```

### 3. 启动服务

```bash
./bin/hft-sim
```

服务启动后：
- API 服务器: http://localhost:8080
- Web Dashboard: http://localhost:8080 (首页)

## 部署方式

### 单机部署

```bash
# 后台运行
nohup ./bin/hft-sim > hft.log 2>&1 &

# 使用 systemd (创建 /etc/systemd/system/hft-sim.service)
```

systemd 服务示例：

```ini
[Unit]
Description=HFT Simulated Exchange
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/ShadowHFT
ExecStart=/home/ubuntu/ShadowHFT/bin/hft-sim
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable hft-sim
sudo systemctl start hft-sim
```

### Docker 部署 (可选)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o hft-sim .
RUN go build -o admin ./cmd/admin

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/hft-sim .
COPY --from=builder /app/admin .
COPY --from=builder /app/web ./web
EXPOSE 8080
CMD ["./hft-sim"]
```

## API 使用示例

### 使用 cURL

```bash
# 查询账户信息
curl -H "X-MBX-APIKEY: <your-api-key>" \
  http://localhost:8080/api/v3/account

# 创建限价买单
curl -X POST http://localhost:8080/api/v3/order \
  -H "X-MBX-APIKEY: <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "side": "BUY",
    "type": "LIMIT",
    "quantity": "0.01",
    "price": "40000",
    "leverage": 10
  }'

# 查询未成交订单
curl -H "X-MBX-APIKEY: <your-api-key>" \
  http://localhost:8080/api/v3/openOrders

# 撤单
curl -X DELETE "http://localhost:8080/api/v3/order?orderId=1" \
  -H "X-MBX-APIKEY: <your-api-key>"
```

### 使用 CCXT (Python)

```python
import ccxt

# 配置模拟交易所
exchange = ccxt.binance({
    'apiKey': '<your-api-key>',
    'secret': 'dummy-secret',  # 模拟交易所不需要验证签名
    'urls': {
        'api': {
            'public': 'http://localhost:8080/api/v3',
            'private': 'http://localhost:8080/api/v3',
        }
    },
    'enableRateLimit': True,
})

# 查询余额
balance = exchange.fetch_balance()
print(balance)

# 创建限价单
order = exchange.create_limit_buy_order('BTC/USDT', 0.01, 40000)
print(order)

# 查询持仓
positions = exchange.fetch_positions()
print(positions)
```

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        币安 WebSocket                        │
│                   wss://stream.binance.com                   │
└───────────────────────────┬─────────────────────────────────┘
                            │ aggTrade 流
┌───────────────────────────▼─────────────────────────────────┐
│                    Data Collector                            │
│              (internal/collector/collector.go)               │
└───────────────────────────┬─────────────────────────────────┘
                            │ Trade 事件
┌───────────────────────────▼─────────────────────────────────┐
│                   Matching Engine                            │
│              (internal/matching/engine.go)                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ Order Store  │  │Position Store│  │Balance Store │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
└───────────────────────────┬─────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│  REST API     │  │  WebSocket    │  │  Web Dashboard│
│  /api/v3/*    │  │  /ws          │  │  /            │
└───────────────┘  └───────────────┘  └───────────────┘
```

## 撮合规则

- **Price-Through 成交**: 买单在价格低于限价时成交，卖单在价格高于限价时成交
- **Maker 费率**: 0.02%
- **杠杆支持**: 1-125 倍（通过配置调整）
- **强平机制**: 维护保证金率 0.5%（TODO）

## 配置项

配置存储在 SQLite 数据库的 `config` 表中，可通过 SQL 修改：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| supported_symbols | ["BTCUSDT","ETHUSDT"] | 支持的交易对 |
| max_leverage | 125 | 最大杠杆倍数 |
| default_leverage | 10 | 默认杠杆倍数 |
| trade_fee_maker | 0.0002 | Maker 手续费率 |
| trade_fee_taker | 0.0005 | Taker 手续费率 |
| binance_ws_url | wss://stream.binance.com:9443/ws | 币安 WebSocket 地址 |

## 数据存储

- **数据库**: SQLite (`hft.db`)
- **表结构**:
  - `api_keys`: 策略账户信息
  - `orders`: 订单记录
  - `trades`: 成交记录
  - `positions`: 持仓信息
  - `balances`: 账户余额
  - `config`: 系统配置

## 后续优化

- [ ] 完善撮合逻辑：部分成交、反向持仓平仓计算
- [ ] 强平机制：监控保证金率，触发自动强平
- [ ] 统计数据：收益率、最大回撤、夏普比率
- [ ] WebSocket 推送：实时推送订单状态和持仓变化
- [ ] 前端完善：连接真实 API，实时更新排行榜
- [ ] 更多测试覆盖

## License

MIT
