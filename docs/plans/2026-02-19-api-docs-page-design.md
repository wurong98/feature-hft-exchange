# API 文档页面设计方案

## 概述

为 HFT 模拟交易所创建一个独立的 API 文档和测试页面 (`/api-docs`)，类似 FastAPI 的 `/docs` 界面。提供完整的 API 端点文档展示和交互式测试功能。

## 设计目标

1. **完整文档展示** - 展示所有 API 端点的请求/响应格式
2. **交互式测试** - 一键发起 API 测试请求
3. **curl 示例** - 自动生成并展示 curl 命令，支持一键复制
4. **用户友好** - 深色主题，清晰的视觉层次

## 页面结构

```
/api-docs (独立页面)
├── 左侧导航栏 (fixed, 280px width)
│   ├── 搜索框 (过滤端点)
│   └── 端点分类列表
│       ├── 账户 API (/api/v3/account)
│       ├── 订单 API (/api/v3/order, /api/v3/openOrders, /api/v3/myTrades)
│       ├── 市场数据 API (/api/v3/exchangeInfo, /api/v3/depth)
│       └── Dashboard API (/api/dashboard/*)
├── 顶部工具栏 (sticky)
│   ├── API Key 输入框（全局，保存到 localStorage）
│   └── 基础 URL 显示 (readonly)
└── 主内容区 (scrollable)
    └── 端点卡片 (每个端点一个)
        ├── 头部: HTTP方法 + 路径 + 描述
        ├── curl 示例区 (带复制按钮)
        ├── 参数表单区
        │   ├── GET: Query 参数输入表格
        │   └── POST: JSON body 编辑器
        ├── [执行请求] 按钮
        └── 响应展示区 (执行后显示)
            ├── 状态码 + 响应时间
            └── 格式化 JSON 响应体
```

## UI 设计规范

### 配色方案 (延续 Dashboard 深色主题)
- 背景色: `#0d1117` (主), `#161b22` (卡片), `#21262d` (输入框)
- 文字色: `#f0f6fc` (主), `#c9d1d9` (次), `#8b949e` ( muted)
- 边框色: `#30363d`
- HTTP 方法色:
  - GET: `#58a6ff` (蓝色)
  - POST: `#3fb950` (绿色)
  - DELETE: `#f85149` (红色)
- 强调色: `#58a6ff`

### 组件样式
- **方法标签**: 圆角矩形背景，带颜色区分
- **端点卡片**: 边框卡片，hover 时高亮
- **输入框**: 深色背景，聚焦时边框变蓝
- **按钮**: 主按钮蓝色背景，次要按钮边框样式
- **代码块**: 等宽字体，语法高亮

## 功能特性

### 1. 左侧导航
- 按 API 类别分组展示端点
- 点击跳转到对应端点卡片
- 搜索框实时过滤端点列表

### 2. 全局 API Key
- 页面顶部固定输入框
- 自动保存到 localStorage
- 所有测试请求共用此 Key

### 3. curl 示例生成
- 根据当前填写的参数实时生成 curl 命令
- 点击复制按钮复制到剪贴板
- 支持自动换行格式化

### 4. 参数表单
- **Query 参数**: 表格形式，每行 (参数名, 类型, 必填, 输入框)
- **Body 参数**: JSON 编辑器，支持语法高亮

### 5. 响应展示
- 显示 HTTP 状态码和响应时间
- JSON 格式化显示，带语法高亮
- 错误响应特殊标记（红色）

## 技术实现

### 前端技术栈
- 原生 HTML/CSS/JavaScript（无框架，保持轻量）
- 内联样式（单文件便于部署）
- 使用 `fetch()` 发送 API 请求

### 文件结构
```
web/
├── index.html      # Dashboard (现有)
├── app.js          # Dashboard JS (现有)
├── api-docs.html   # API 文档页面 (新建)
└── api-docs.js     # API 文档 JS (新建)
```

### 后端路由
- 在 Gin 中添加 `/api-docs` 路由
- 返回 `web/api-docs.html`

## API 端点清单

### 账户 API
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v3/account | 获取账户信息 |

### 订单 API
| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v3/order | 创建订单 |
| DELETE | /api/v3/order | 取消订单 |
| GET | /api/v3/openOrders | 获取未成交订单 |
| GET | /api/v3/myTrades | 获取成交历史 |

### 市场数据 API
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v3/exchangeInfo | 交易所信息 |
| GET | /api/v3/depth | 订单簿深度 |

### Dashboard API
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/dashboard/leaderboard | 排行榜 |
| GET | /api/dashboard/strategy/:apiKey | 策略详情 |
| GET | /api/dashboard/strategy/:apiKey/trades | 策略成交 |
| GET | /api/dashboard/strategy/:apiKey/positions | 策略持仓 |
| GET | /api/dashboard/strategy/:apiKey/orders | 策略委托 |
| GET | /api/dashboard/strategy/:apiKey/snapshots | 收益快照 |
| GET | /api/orderbook/:symbol | 订单簿数据 |

## 测试示例

### 创建订单请求示例
```json
{
  "symbol": "BTCUSDT",
  "side": "BUY",
  "type": "LIMIT",
  "quantity": "0.01",
  "price": "65000",
  "leverage": 10
}
```

### 生成的 curl 示例
```bash
curl -X POST 'http://localhost:8080/api/v3/order' \
  -H 'X-MBX-APIKEY: YOUR_API_KEY' \
  -H 'Content-Type: application/json' \
  -d '{
    "symbol": "BTCUSDT",
    "side": "BUY",
    "type": "LIMIT",
    "quantity": "0.01",
    "price": "65000",
    "leverage": 10
  }'
```

## 后续优化

- [ ] 支持 WebSocket 测试
- [ ] 响应数据类型验证
- [ ] 导入/导出 API 请求集合
- [ ] 生成多语言代码示例 (Python, Go, JavaScript)
