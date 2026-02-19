# API 文档页面实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 创建一个独立的 API 文档和测试页面 (`/api-docs`)，提供完整的 API 文档展示和交互式测试功能。

**Architecture:** 独立的 HTML/JS 页面，左侧导航栏展示 API 端点列表，主内容区显示每个端点的详细信息、curl 示例、参数表单和响应展示。使用原生 JavaScript 发送 fetch 请求进行测试。

**Tech Stack:** HTML5, CSS3 (Flexbox/Grid), Vanilla JavaScript, Gin (后端路由)

---

### Task 1: 创建 API 文档页面 HTML 结构

**Files:**
- Create: `web/api-docs.html`

**Step 1: 创建基础 HTML 结构**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HFT API 文档</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }

        :root {
            --bg-primary: #0d1117;
            --bg-secondary: #161b22;
            --bg-tertiary: #21262d;
            --border-color: #30363d;
            --text-primary: #f0f6fc;
            --text-secondary: #c9d1d9;
            --text-muted: #8b949e;
            --color-accent: #58a6ff;
            --color-get: #58a6ff;
            --color-post: #3fb950;
            --color-delete: #f85149;
            --color-long: #3fb950;
            --color-short: #f85149;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg-primary);
            color: var(--text-secondary);
            font-size: 14px;
            line-height: 1.5;
        }

        /* Header */
        .header {
            background: var(--bg-secondary);
            border-bottom: 1px solid var(--border-color);
            padding: 12px 20px;
            position: sticky;
            top: 0;
            z-index: 100;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .header-brand h1 {
            color: var(--color-accent);
            font-size: 18px;
            font-weight: 600;
        }

        .header-tools {
            display: flex;
            align-items: center;
            gap: 16px;
        }

        .api-key-input {
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .api-key-input label {
            color: var(--text-muted);
            font-size: 12px;
        }

        .api-key-input input {
            background: var(--bg-tertiary);
            border: 1px solid var(--border-color);
            border-radius: 4px;
            padding: 6px 10px;
            color: var(--text-primary);
            font-size: 13px;
            width: 280px;
        }

        .api-key-input input:focus {
            outline: none;
            border-color: var(--color-accent);
        }

        /* Main Layout */
        .main-container {
            display: flex;
            height: calc(100vh - 50px);
        }

        /* Sidebar */
        .sidebar {
            width: 280px;
            background: var(--bg-secondary);
            border-right: 1px solid var(--border-color);
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }

        .sidebar-header {
            padding: 16px;
            border-bottom: 1px solid var(--border-color);
        }

        .search-box {
            position: relative;
        }

        .search-box input {
            width: 100%;
            background: var(--bg-tertiary);
            border: 1px solid var(--border-color);
            border-radius: 6px;
            padding: 8px 12px;
            color: var(--text-secondary);
            font-size: 13px;
            outline: none;
        }

        .search-box input:focus {
            border-color: var(--color-accent);
        }

        .sidebar-content {
            flex: 1;
            overflow-y: auto;
            padding: 8px 0;
        }

        .nav-category {
            margin-bottom: 8px;
        }

        .nav-category-title {
            padding: 8px 16px;
            color: var(--text-muted);
            font-size: 11px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .nav-item {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 8px 16px;
            cursor: pointer;
            transition: background 0.2s;
        }

        .nav-item:hover {
            background: var(--bg-tertiary);
        }

        .nav-method {
            font-size: 11px;
            font-weight: 600;
            padding: 2px 6px;
            border-radius: 3px;
            min-width: 50px;
            text-align: center;
        }

        .nav-method.get { background: rgba(88, 166, 255, 0.2); color: var(--color-get); }
        .nav-method.post { background: rgba(63, 185, 80, 0.2); color: var(--color-post); }
        .nav-method.delete { background: rgba(248, 81, 73, 0.2); color: var(--color-delete); }

        .nav-path {
            font-size: 12px;
            color: var(--text-secondary);
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        /* Content Area */
        .content {
            flex: 1;
            overflow-y: auto;
            padding: 20px;
        }

        /* Endpoint Card */
        .endpoint-card {
            background: var(--bg-secondary);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            margin-bottom: 16px;
            overflow: hidden;
        }

        .endpoint-header {
            padding: 16px;
            border-bottom: 1px solid var(--border-color);
            display: flex;
            align-items: center;
            gap: 12px;
        }

        .endpoint-method {
            font-size: 12px;
            font-weight: 600;
            padding: 4px 10px;
            border-radius: 4px;
        }

        .endpoint-method.get { background: rgba(88, 166, 255, 0.2); color: var(--color-get); }
        .endpoint-method.post { background: rgba(63, 185, 80, 0.2); color: var(--color-post); }
        .endpoint-method.delete { background: rgba(248, 81, 73, 0.2); color: var(--color-delete); }

        .endpoint-path {
            font-family: 'SF Mono', Monaco, monospace;
            font-size: 14px;
            color: var(--text-primary);
        }

        .endpoint-desc {
            margin-left: auto;
            color: var(--text-muted);
            font-size: 13px;
        }

        /* Curl Section */
        .section {
            padding: 16px;
            border-bottom: 1px solid var(--border-color);
        }

        .section:last-child {
            border-bottom: none;
        }

        .section-title {
            font-size: 12px;
            font-weight: 600;
            color: var(--text-primary);
            margin-bottom: 12px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .code-block {
            background: var(--bg-primary);
            border: 1px solid var(--border-color);
            border-radius: 6px;
            padding: 12px;
            font-family: 'SF Mono', Monaco, monospace;
            font-size: 12px;
            overflow-x: auto;
            position: relative;
        }

        .code-block pre {
            margin: 0;
            white-space: pre-wrap;
            word-break: break-all;
            color: var(--text-secondary);
        }

        .copy-btn {
            position: absolute;
            top: 8px;
            right: 8px;
            background: var(--bg-tertiary);
            border: 1px solid var(--border-color);
            border-radius: 4px;
            padding: 4px 10px;
            color: var(--text-muted);
            font-size: 11px;
            cursor: pointer;
            transition: all 0.2s;
        }

        .copy-btn:hover {
            background: var(--border-color);
            color: var(--text-primary);
        }

        /* Parameter Form */
        .param-table {
            width: 100%;
            border-collapse: collapse;
            font-size: 13px;
        }

        .param-table th {
            text-align: left;
            padding: 10px;
            color: var(--text-muted);
            font-weight: 500;
            border-bottom: 1px solid var(--border-color);
        }

        .param-table td {
            padding: 10px;
            border-bottom: 1px solid var(--border-color);
        }

        .param-table tr:last-child td {
            border-bottom: none;
        }

        .param-name {
            font-family: 'SF Mono', Monaco, monospace;
            color: var(--text-primary);
        }

        .param-type {
            color: var(--text-muted);
            font-size: 12px;
        }

        .param-required {
            color: var(--color-short);
            font-size: 11px;
        }

        .param-input {
            background: var(--bg-tertiary);
            border: 1px solid var(--border-color);
            border-radius: 4px;
            padding: 6px 10px;
            color: var(--text-primary);
            font-size: 13px;
            width: 100%;
        }

        .param-input:focus {
            outline: none;
            border-color: var(--color-accent);
        }

        textarea.param-input {
            min-height: 120px;
            font-family: 'SF Mono', Monaco, monospace;
            font-size: 12px;
            resize: vertical;
        }

        /* Execute Button */
        .execute-btn {
            background: var(--color-accent);
            color: white;
            border: none;
            border-radius: 6px;
            padding: 10px 24px;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            transition: opacity 0.2s;
        }

        .execute-btn:hover {
            opacity: 0.9;
        }

        /* Response Section */
        .response-meta {
            display: flex;
            align-items: center;
            gap: 16px;
            margin-bottom: 12px;
        }

        .response-status {
            font-size: 13px;
            font-weight: 600;
        }

        .response-status.success { color: var(--color-long); }
        .response-status.error { color: var(--color-short); }

        .response-time {
            color: var(--text-muted);
            font-size: 12px;
        }

        .json-key { color: #7ee787; }
        .json-string { color: #a5d6ff; }
        .json-number { color: #79c0ff; }
        .json-boolean { color: #ff7b72; }
        .json-null { color: #ff7b72; }
    </style>
</head>
<body>
    <header class="header">
        <div class="header-brand">
            <h1>HFT API 文档</h1>
        </div>
        <div class="header-tools">
            <div class="api-key-input">
                <label>API Key:</label>
                <input type="text" id="api-key" placeholder="输入您的 API Key">
            </div>
        </div>
    </header>

    <div class="main-container">
        <aside class="sidebar">
            <div class="sidebar-header">
                <div class="search-box">
                    <input type="text" id="search-input" placeholder="搜索 API...">
                </div>
            </div>
            <div class="sidebar-content" id="nav-content">
                <!-- Navigation items will be generated by JS -->
            </div>
        </aside>

        <main class="content" id="main-content">
            <!-- Endpoint cards will be generated by JS -->
        </main>
    </div>

    <script src="api-docs.js"></script>
</body>
</html>
```

**Step 2: Commit**

```bash
git add web/api-docs.html
git commit -m "feat: add API docs page HTML structure"
```

---

### Task 2: 创建 API 文档 JavaScript 逻辑

**Files:**
- Create: `web/api-docs.js`

**Step 1: 定义 API 端点配置**

```javascript
const API_BASE = window.location.origin;

const API_ENDPOINTS = [
    {
        category: '账户 API',
        endpoints: [
            {
                method: 'GET',
                path: '/api/v3/account',
                desc: '获取账户信息，包括余额和持仓',
                auth: true,
                params: []
            }
        ]
    },
    {
        category: '订单 API',
        endpoints: [
            {
                method: 'POST',
                path: '/api/v3/order',
                desc: '创建限价单',
                auth: true,
                body: {
                    symbol: { type: 'string', required: true, default: 'BTCUSDT', desc: '交易对' },
                    side: { type: 'string', required: true, default: 'BUY', desc: '方向: BUY/SELL' },
                    type: { type: 'string', required: true, default: 'LIMIT', desc: '类型: LIMIT' },
                    quantity: { type: 'string', required: true, default: '0.01', desc: '数量' },
                    price: { type: 'string', required: true, default: '65000', desc: '价格' },
                    leverage: { type: 'integer', required: false, default: 10, desc: '杠杆倍数' }
                }
            },
            {
                method: 'DELETE',
                path: '/api/v3/order',
                desc: '取消订单',
                auth: true,
                params: [
                    { name: 'orderId', type: 'string', required: true, default: '', desc: '订单ID' },
                    { name: 'symbol', type: 'string', required: true, default: 'BTCUSDT', desc: '交易对' }
                ]
            },
            {
                method: 'GET',
                path: '/api/v3/openOrders',
                desc: '获取所有未成交订单',
                auth: true,
                params: [
                    { name: 'symbol', type: 'string', required: false, default: '', desc: '交易对(可选)' }
                ]
            },
            {
                method: 'GET',
                path: '/api/v3/myTrades',
                desc: '获取成交历史',
                auth: true,
                params: [
                    { name: 'symbol', type: 'string', required: false, default: '', desc: '交易对(可选)' }
                ]
            }
        ]
    },
    {
        category: '市场数据 API',
        endpoints: [
            {
                method: 'GET',
                path: '/api/v3/exchangeInfo',
                desc: '获取交易所信息和交易对列表',
                auth: false,
                params: []
            },
            {
                method: 'GET',
                path: '/api/v3/depth',
                desc: '获取订单簿深度',
                auth: false,
                params: [
                    { name: 'symbol', type: 'string', required: false, default: 'BTCUSDT', desc: '交易对' }
                ]
            }
        ]
    },
    {
        category: 'Dashboard API',
        endpoints: [
            {
                method: 'GET',
                path: '/api/dashboard/leaderboard',
                desc: '获取策略排行榜',
                auth: false,
                params: []
            },
            {
                method: 'GET',
                path: '/api/dashboard/strategy/{apiKey}',
                desc: '获取策略详情',
                auth: false,
                params: [
                    { name: 'apiKey', type: 'string', required: true, default: '', desc: '策略API Key' }
                ]
            },
            {
                method: 'GET',
                path: '/api/dashboard/strategy/{apiKey}/trades',
                desc: '获取策略成交记录',
                auth: false,
                params: [
                    { name: 'apiKey', type: 'string', required: true, default: '', desc: '策略API Key' }
                ]
            },
            {
                method: 'GET',
                path: '/api/dashboard/strategy/{apiKey}/positions',
                desc: '获取策略持仓',
                auth: false,
                params: [
                    { name: 'apiKey', type: 'string', required: true, default: '', desc: '策略API Key' }
                ]
            },
            {
                method: 'GET',
                path: '/api/dashboard/strategy/{apiKey}/orders',
                desc: '获取策略当前委托',
                auth: false,
                params: [
                    { name: 'apiKey', type: 'string', required: true, default: '', desc: '策略API Key' }
                ]
            },
            {
                method: 'GET',
                path: '/api/dashboard/strategy/{apiKey}/snapshots',
                desc: '获取策略收益快照',
                auth: false,
                params: [
                    { name: 'apiKey', type: 'string', required: true, default: '', desc: '策略API Key' },
                    { name: 'limit', type: 'integer', required: false, default: '100', desc: '返回条数' }
                ]
            },
            {
                method: 'GET',
                path: '/api/orderbook/{symbol}',
                desc: '获取订单簿数据',
                auth: false,
                params: [
                    { name: 'symbol', type: 'string', required: true, default: 'BTCUSDT', desc: '交易对' }
                ]
            }
        ]
    }
];
```

**Step 2: 添加 DOM 生成函数**

```javascript
function generateCurlCommand(method, path, params, body, apiKey) {
    let url = `${API_BASE}${path}`;
    let curl = `curl -X ${method} '${url}'`;

    if (apiKey) {
        curl += ` \\\n  -H 'X-MBX-APIKEY: ${apiKey}'`;
    }

    if (method === 'POST' && body) {
        curl += ` \\\n  -H 'Content-Type: application/json'`;
        curl += ` \\\n  -d '${JSON.stringify(body, null, 2)}'`;
    }

    if (params && params.length > 0) {
        const queryParams = params.filter(p => p.value).map(p => `${p.name}=${encodeURIComponent(p.value)}`).join('&');
        if (queryParams) {
            url += `?${queryParams}`;
            curl = `curl -X ${method} '${url}'`;
            if (apiKey) {
                curl += ` \\\n  -H 'X-MBX-APIKEY: ${apiKey}'`;
            }
        }
    }

    return curl;
}

function syntaxHighlight(json) {
    if (!json) return '';
    const str = typeof json === 'string' ? json : JSON.stringify(json, null, 2);
    return str
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/(".*?"):/g, '<span class="json-key">$1</span>:')
        .replace(/: (".*?")/g, ': <span class="json-string">$1</span>')
        .replace(/: (\d+\.?\d*)/g, ': <span class="json-number">$1</span>')
        .replace(/: (true|false)/g, ': <span class="json-boolean">$1</span>')
        .replace(/: (null)/g, ': <span class="json-null">$1</span>');
}

function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        alert('已复制到剪贴板');
    });
}
```

**Step 3: 添加渲染函数**

```javascript
function renderNavigation() {
    const navContent = document.getElementById('nav-content');
    let html = '';

    API_ENDPOINTS.forEach(category => {
        html += `
            <div class="nav-category">
                <div class="nav-category-title">${category.category}</div>
        `;

        category.endpoints.forEach(endpoint => {
            const endpointId = `${endpoint.method}_${endpoint.path.replace(/[^a-zA-Z0-9]/g, '_')}`;
            html += `
                <div class="nav-item" onclick="scrollToEndpoint('${endpointId}')">
                    <span class="nav-method ${endpoint.method.toLowerCase()}">${endpoint.method}</span>
                    <span class="nav-path">${endpoint.path}</span>
                </div>
            `;
        });

        html += '</div>';
    });

    navContent.innerHTML = html;
}

function renderEndpoints() {
    const mainContent = document.getElementById('main-content');
    let html = '';

    API_ENDPOINTS.forEach(category => {
        category.endpoints.forEach(endpoint => {
            const endpointId = `${endpoint.method}_${endpoint.path.replace(/[^a-zA-Z0-9]/g, '_')}`;
            html += `
                <div class="endpoint-card" id="${endpointId}">
                    <div class="endpoint-header">
                        <span class="endpoint-method ${endpoint.method.toLowerCase()}">${endpoint.method}</span>
                        <span class="endpoint-path">${endpoint.path}</span>
                        <span class="endpoint-desc">${endpoint.desc}</span>
                    </div>

                    <div class="section">
                        <div class="section-title">curl 示例</div>
                        <div class="code-block">
                            <button class="copy-btn" onclick="copyCurl('${endpointId}')">复制</button>
                            <pre id="curl-${endpointId}"></pre>
                        </div>
                    </div>
            `;

            // Parameters section
            if (endpoint.params && endpoint.params.length > 0) {
                html += `
                    <div class="section">
                        <div class="section-title">Query 参数</div>
                        <table class="param-table">
                            <tr>
                                <th>参数名</th>
                                <th>类型</th>
                                <th>必填</th>
                                <th>值</th>
                            </tr>
                `;

                endpoint.params.forEach(param => {
                    html += `
                        <tr>
                            <td><span class="param-name">${param.name}</span></td>
                            <td><span class="param-type">${param.type}</span></td>
                            <td>${param.required ? '<span class="param-required">是</span>' : '否'}</td>
                            <td>
                                <input type="text" class="param-input"
                                    id="param-${endpointId}-${param.name}"
                                    value="${param.default || ''}"
                                    placeholder="${param.desc}">
                            </td>
                        </tr>
                    `;
                });

                html += '</table></div>';
            }

            // Body section
            if (endpoint.body) {
                const bodyDefaults = {};
                Object.entries(endpoint.body).forEach(([key, value]) => {
                    bodyDefaults[key] = value.default;
                });

                html += `
                    <div class="section">
                        <div class="section-title">请求体 (JSON)</div>
                        <textarea class="param-input" id="body-${endpointId}"
                            placeholder="请求体 JSON">${JSON.stringify(bodyDefaults, null, 2)}</textarea>
                    </div>
                `;
            }

            // Execute button
            html += `
                    <div class="section">
                        <button class="execute-btn" onclick="executeRequest('${endpointId}', '${endpoint.method}', '${endpoint.path}')">
                            执行请求
                        </button>
                    </div>

                    <div class="section" id="response-${endpointId}" style="display: none;">
                        <div class="section-title">响应</div>
                        <div class="response-meta">
                            <span class="response-status" id="status-${endpointId}"></span>
                            <span class="response-time" id="time-${endpointId}"></span>
                        </div>
                        <div class="code-block">
                            <pre id="response-body-${endpointId}"></pre>
                        </div>
                    </div>
                </div>
            `;
        });
    });

    mainContent.innerHTML = html;

    // Update curl commands
    API_ENDPOINTS.forEach(category => {
        category.endpoints.forEach(endpoint => {
            updateCurlCommand(endpoint);
        });
    });
}
```

**Step 4: 添加交互函数**

```javascript
function scrollToEndpoint(endpointId) {
    document.getElementById(endpointId).scrollIntoView({ behavior: 'smooth' });
}

function copyCurl(endpointId) {
    const curlText = document.getElementById(`curl-${endpointId}`).textContent;
    copyToClipboard(curlText);
}

function updateCurlCommand(endpoint) {
    const endpointId = `${endpoint.method}_${endpoint.path.replace(/[^a-zA-Z0-9]/g, '_')}`;
    const apiKey = document.getElementById('api-key').value;

    let params = [];
    if (endpoint.params) {
        params = endpoint.params.map(p => ({
            name: p.name,
            value: document.getElementById(`param-${endpointId}-${p.name}`)?.value || p.default
        }));
    }

    let body = null;
    if (endpoint.body) {
        try {
            body = JSON.parse(document.getElementById(`body-${endpointId}`)?.value || '{}');
        } catch (e) {
            body = {};
        }
    }

    const curl = generateCurlCommand(endpoint.method, endpoint.path, params, body, apiKey);
    document.getElementById(`curl-${endpointId}`).textContent = curl;
}

async function executeRequest(endpointId, method, path) {
    const apiKey = document.getElementById('api-key').value;
    const responseSection = document.getElementById(`response-${endpointId}`);
    const statusEl = document.getElementById(`status-${endpointId}`);
    const timeEl = document.getElementById(`time-${endpointId}`);
    const bodyEl = document.getElementById(`response-body-${endpointId}`);

    const endpoint = API_ENDPOINTS.flatMap(c => c.endpoints).find(e =>
        `${e.method}_${e.path.replace(/[^a-zA-Z0-9]/g, '_')}` === endpointId
    );

    // Build URL with query params
    let url = `${API_BASE}${path}`;
    if (endpoint.params) {
        const queryParams = endpoint.params
            .map(p => ({
                name: p.name,
                value: document.getElementById(`param-${endpointId}-${p.name}`)?.value
            }))
            .filter(p => p.value)
            .map(p => `${p.name}=${encodeURIComponent(p.value)}`)
            .join('&');
        if (queryParams) {
            url += `?${queryParams}`;
        }
    }

    // Replace path params
    endpoint.params?.forEach(p => {
        const value = document.getElementById(`param-${endpointId}-${p.name}`)?.value;
        if (value) {
            url = url.replace(`{${p.name}}`, value);
        }
    });

    const headers = {
        'Content-Type': 'application/json'
    };
    if (apiKey) {
        headers['X-MBX-APIKEY'] = apiKey;
    }

    const options = {
        method: method,
        headers: headers
    };

    if (method === 'POST' && endpoint.body) {
        try {
            options.body = document.getElementById(`body-${endpointId}`).value;
        } catch (e) {
            alert('请求体 JSON 格式错误');
            return;
        }
    }

    const startTime = performance.now();

    try {
        const response = await fetch(url, options);
        const duration = Math.round(performance.now() - startTime);

        let data;
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            data = await response.json();
        } else {
            data = await response.text();
        }

        statusEl.textContent = `${response.status} ${response.statusText}`;
        statusEl.className = `response-status ${response.ok ? 'success' : 'error'}`;
        timeEl.textContent = `${duration}ms`;
        bodyEl.innerHTML = syntaxHighlight(data);
        responseSection.style.display = 'block';
    } catch (error) {
        statusEl.textContent = 'Error';
        statusEl.className = 'response-status error';
        timeEl.textContent = '';
        bodyEl.textContent = error.message;
        responseSection.style.display = 'block';
    }
}
```

**Step 5: 初始化**

```javascript
document.addEventListener('DOMContentLoaded', () => {
    // Load API Key from localStorage
    const savedApiKey = localStorage.getItem('hft_api_key');
    if (savedApiKey) {
        document.getElementById('api-key').value = savedApiKey;
    }

    // Save API Key to localStorage on change
    document.getElementById('api-key').addEventListener('input', (e) => {
        localStorage.setItem('hft_api_key', e.target.value);
        // Update all curl commands
        API_ENDPOINTS.forEach(category => {
            category.endpoints.forEach(updateCurlCommand);
        });
    });

    // Search functionality
    document.getElementById('search-input').addEventListener('input', (e) => {
        const searchTerm = e.target.value.toLowerCase();
        document.querySelectorAll('.nav-item').forEach(item => {
            const text = item.textContent.toLowerCase();
            item.style.display = text.includes(searchTerm) ? 'flex' : 'none';
        });
    });

    renderNavigation();
    renderEndpoints();
});
```

**Step 6: Commit**

```bash
git add web/api-docs.js
git commit -m "feat: add API docs JavaScript logic"
```

---

### Task 3: 添加后端路由

**Files:**
- Modify: `internal/api/server.go:40-60` (路由注册区域)

**Step 1: 添加 API docs 路由**

在 `setupRoutes()` 函数中，添加 `/api-docs` 路由：

```go
// API Docs page
r.GET("/api-docs", func(c *gin.Context) {
    c.File("./web/api-docs.html")
})
```

放置在 `/dashboard` 路由附近：

```go
// Dashboard
r.GET("/dashboard", func(c *gin.Context) {
    c.File("./web/index.html")
})

// API Docs page
r.GET("/api-docs", func(c *gin.Context) {
    c.File("./web/api-docs.html")
})
```

**Step 2: Commit**

```bash
git add internal/api/server.go
git commit -m "feat: add /api-docs route"
```

---

### Task 4: 测试验证

**Step 1: 构建并启动服务**

```bash
go build -o bin/hft-sim .
./bin/hft-sim
```

**Step 2: 访问 API 文档页面**

打开浏览器访问: `http://localhost:8080/api-docs`

**Step 3: 验证功能**

1. 页面加载是否正常，左侧导航栏是否显示所有 API 端点
2. 搜索框是否可以过滤端点
3. 点击左侧导航项是否能跳转到对应端点卡片
4. 输入 API Key 后 curl 示例是否正确更新
5. 点击"执行请求"按钮是否能发送请求并显示响应
6. 复制 curl 按钮是否正常工作

**Step 4: Commit (如果测试通过)**

```bash
git commit -m "test: verify api docs page functionality" --allow-empty
```

---

## 快速测试命令

```bash
# 构建
go build -o bin/hft-sim .

# 启动
./bin/hft-sim

# 访问
open http://localhost:8080/api-docs
```
