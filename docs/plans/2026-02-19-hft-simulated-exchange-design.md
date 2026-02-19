# 模拟交易所系统设计文档

## 1. 项目概述

一个基于真实币安 trade 数据流驱动的永续合约模拟交易所系统。用户通过 CCXT 兼容 API 提交限价单，系统根据币安实时成交数据判断价格击穿并撮合，提供收益分析和公开排行榜。

### 核心特性
- 基于币安 WebSocket trade 流的实时撮合
- Price-through 成交判定（买单：trade.price < limit_price；卖单：trade.price > limit_price）
- 只支持限价单（LIMIT）
- CCXT 兼容 API，用户无缝接入
- 永续合约支持（全仓、杠杆可配置）
- 策略公开透明（所有人可见挂单、成交、盈亏）
- 邀请码机制获取 API Key

## 2. 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                     币安 WebSocket                        │
│                    (trade 数据流)                        │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                  模拟交易所系统 (Go)                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │ Data         │  │ Matching     │  │ API Server   │   │
│  │ Collector    │──│ Engine       │──│ (CCXT兼容)   │   │
│  └──────────────┘  └──────────────┘  └──────────────┘   │
│         │                   │                │          │
│         └───────────────────┴────────────────┘          │
│                             │                           │
│                      ┌──────▼──────┐                    │
│                      │   SQLite    │                    │
│                      │  (orders,    │                    │
│                      │  positions,  │                    │
│                      │  trades)     │                    │
│                      └─────────────┘                    │
│                             │                           │
│                      ┌──────▼──────┐                    │
│                      │   Web       │                    │
│                      │ Dashboard   │                    │
│                      │(策略排行榜) │                    │
│                      └─────────────┘                    │
└─────────────────────────────────────────────────────────┘
          ▲                                   ▲
          │                                   │
    ┌─────┴─────┐                       ┌─────┴─────┐
    │  用户策略  │                       │   站长    │
    │ (CCXT API)│                       │(邀请码管理)│
    └───────────┘                       └───────────┘
```

### 组件说明

#### 2.1 Data Collector（数据收集器）
- 连接币安 `wss://stream.binance.com:9443/ws/<symbol>@aggTrade`
- 支持多交易对同时订阅（通过配置）
- 将 trade 数据推送到内部 channel

#### 2.2 Matching Engine（撮合引擎）
- 接收 trade 数据
- 遍历该 symbol 的所有未成交限价单
- 判断 price-through 条件：
  - 买单：trade.price < limit_price
  - 卖单：trade.price > limit_price
- 成交后更新持仓、计算盈亏、记录历史

#### 2.3 API Server（CCXT 兼容）
提供以下端点：
- `POST /api/v3/order` - 创建限价单
- `DELETE /api/v3/order` - 取消订单
- `GET /api/v3/openOrders` - 查询未成交订单
- `GET /api/v3/account` - 查询账户信息
- `GET /api/v3/myTrades` - 查询成交历史
- `WebSocket /ws/stream` - 实时推送

#### 2.4 Web Dashboard
- 策略排行榜（按收益率排序）
- 策略详情页（资金曲线、持仓、成交记录）
- 实时行情看板

## 3. 数据库设计

```sql
-- 邀请码/策略管理
CREATE TABLE api_keys (
    key TEXT PRIMARY KEY,              -- API Key
    name TEXT NOT NULL,                -- 策略名称
    description TEXT,                  -- 策略描述（可包含广告、联系信息）
    initial_balance DECIMAL NOT NULL,  -- 初始资金（USDT）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 配置表（站长可配置）
CREATE TABLE config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- 订单表
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    api_key TEXT NOT NULL,
    symbol TEXT NOT NULL,
    side TEXT NOT NULL CHECK (side IN ('BUY', 'SELL')),
    type TEXT NOT NULL CHECK (type = 'LIMIT'),
    price DECIMAL NOT NULL,
    quantity DECIMAL NOT NULL,
    executed_qty DECIMAL DEFAULT 0,
    leverage INTEGER DEFAULT 1,
    status TEXT DEFAULT 'NEW' CHECK (status IN ('NEW', 'PARTIALLY_FILLED', 'FILLED', 'CANCELLED')),
    client_order_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (api_key) REFERENCES api_keys(key)
);

CREATE INDEX idx_orders_api_key ON orders(api_key);
CREATE INDEX idx_orders_symbol_status ON orders(symbol, status);

-- 成交记录
CREATE TABLE trades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    api_key TEXT NOT NULL,
    symbol TEXT NOT NULL,
    side TEXT NOT NULL,
    price DECIMAL NOT NULL,
    quantity DECIMAL NOT NULL,
    quote_qty DECIMAL NOT NULL,
    fee DECIMAL NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (api_key) REFERENCES api_keys(key)
);

CREATE INDEX idx_trades_api_key ON trades(api_key);
CREATE INDEX idx_trades_symbol ON trades(symbol);

-- 持仓表（永续合约）
CREATE TABLE positions (
    api_key TEXT NOT NULL,
    symbol TEXT NOT NULL,
    side TEXT NOT NULL CHECK (side IN ('LONG', 'SHORT')),
    entry_price DECIMAL NOT NULL,
    size DECIMAL NOT NULL,
    leverage INTEGER NOT NULL,
    margin DECIMAL NOT NULL,
    unrealized_pnl DECIMAL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (api_key, symbol),
    FOREIGN KEY (api_key) REFERENCES api_keys(key)
);

-- 账户余额
CREATE TABLE balances (
    api_key TEXT PRIMARY KEY,
    available DECIMAL NOT NULL,    -- 可用余额
    frozen DECIMAL DEFAULT 0,      -- 冻结保证金
    total_pnl DECIMAL DEFAULT 0,   -- 累计盈亏
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (api_key) REFERENCES api_keys(key)
);
```

## 4. CCXT 兼容性

### 4.1 认证方式
```
Header: X-MBX-APIKEY: <api_key>
```

### 4.2 API 端点

| 端点 | 方法 | CCXT 方法 |
|------|------|-----------|
| `/api/v3/account` | GET | fetchBalance() |
| `/api/v3/order` | POST | createOrder() |
| `/api/v3/order` | DELETE | cancelOrder() |
| `/api/v3/openOrders` | GET | fetchOpenOrders() |
| `/api/v3/myTrades` | GET | fetchMyTrades() |

### 4.3 订单创建请求格式
```json
{
  "symbol": "BTCUSDT",
  "side": "BUY",
  "type": "LIMIT",
  "timeInForce": "GTC",
  "quantity": "0.01",
  "price": "50000",
  "leverage": 10
}
```

### 4.4 接入示例
```python
import ccxt

exchange = ccxt.binance({
    'apiKey': 'your-api-key',
    'secret': 'not-used',
    'urls': {
        'api': {
            'public': 'http://localhost:8080/api',
            'private': 'http://localhost:8080/api',
        }
    },
    'options': {
        'defaultType': 'future',
    }
})
```

## 5. 永续合约机制

### 5.1 保证金模式
- 全仓模式（简化实现）
- 占用保证金 = 名义价值 / 杠杆

### 5.2 强平逻辑
- 维持保证金率：0.5%（可配置）
- 当保证金率 < 维持保证金率时触发强平
- 强平时按市价（当前 trade 价格）平仓

### 5.3 手续费
- Maker：0.02%
- Taker：0.05%

## 6. 收益统计

### 6.1 实时指标
- 累计收益率 = (当前余额 - 初始资金) / 初始资金 × 100%
- 最大回撤（MDD）
- 夏普比率
- 胜率
- 盈亏比
- 交易频率

### 6.2 公开数据
- 所有人可见：策略名称、描述、收益率、排名、成交记录、持仓
- 仅 API Key 持有者可见：API Key、余额详情

## 7. 配置项

```yaml
# config 表示例值
supported_symbols: "[\"BTCUSDT\",\"ETHUSDT\",\"SOLUSDT\"]"
max_leverage: "125"
default_leverage: "10"
maintenance_margin_rate: "0.005"
trade_fee_maker: "0.0002"
trade_fee_taker: "0.0005"
binance_ws_url: "wss://stream.binance.com:9443/ws"
max_orders_per_api_key: "100"
order_expire_hours: "168"
```

## 8. 错误处理

- WebSocket 断线：自动重连，指数退避
- 撮合异常：记录日志，不影响其他策略
- 余额不足：返回错误码 `-2010`
- 不支持市价单：返回错误码 `-1106`

## 9. 技术栈

- **语言**：Go 1.21+
- **数据库**：SQLite 3
- **Web 框架**：Gin
- **WebSocket**：gorilla/websocket
- **前端**：简单 HTML + JS（可选 Vue/React）

---

生成日期：2026-02-19
