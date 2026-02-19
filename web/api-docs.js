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

    // Replace path params first
    endpoint.params?.forEach(p => {
        const value = document.getElementById(`param-${endpointId}-${p.name}`)?.value;
        if (value) {
            url = url.replace(`{${p.name}}`, value);
        }
    });

    // Add query params for GET requests
    if (method === 'GET' && endpoint.params) {
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

// Setup parameter change listeners to update curl commands
function setupParamListeners() {
    API_ENDPOINTS.forEach(category => {
        category.endpoints.forEach(endpoint => {
            const endpointId = `${endpoint.method}_${endpoint.path.replace(/[^a-zA-Z0-9]/g, '_')}`;

            if (endpoint.params) {
                endpoint.params.forEach(param => {
                    const input = document.getElementById(`param-${endpointId}-${param.name}`);
                    if (input) {
                        input.addEventListener('input', () => updateCurlCommand(endpoint));
                    }
                });
            }

            if (endpoint.body) {
                const textarea = document.getElementById(`body-${endpointId}`);
                if (textarea) {
                    textarea.addEventListener('input', () => updateCurlCommand(endpoint));
                }
            }
        });
    });
}

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

    // Setup listeners after rendering
    setTimeout(setupParamListeners, 0);
});
