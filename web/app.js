// State Management
let currentStrategy = null;
let leaderboardData = [];
let refreshInterval = null;

const API_BASE = '';

// Utility Functions
function formatNumber(num, decimals = 2) {
    if (num === undefined || num === null || num === '') return '-';
    const n = parseFloat(num);
    if (isNaN(n)) return '-';

    if (Math.abs(n) >= 1000000) {
        return (n / 1000000).toFixed(2) + 'M';
    } else if (Math.abs(n) >= 1000) {
        return (n / 1000).toFixed(2) + 'K';
    }
    return n.toFixed(decimals);
}

function formatPrice(price) {
    const p = parseFloat(price);
    if (p >= 1000) return p.toFixed(2);
    if (p >= 1) return p.toFixed(4);
    return p.toFixed(6);
}

function formatTime(timeStr) {
    if (!timeStr) return '-';
    const date = new Date(timeStr);
    const now = new Date();
    const diff = now - date;

    if (diff < 60000) return '刚刚';
    if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前';
    if (diff < 86400000) return Math.floor(diff / 3600000) + '小时前';

    return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}

function getROIColor(roi) {
    const r = parseFloat(roi) || 0;
    return r >= 0 ? 'positive' : 'negative';
}

// Load Leaderboard
async function loadLeaderboard() {
    try {
        const response = await fetch(`${API_BASE}/api/dashboard/leaderboard`);
        if (!response.ok) throw new Error('Failed to load');

        leaderboardData = await response.json();
        renderLeaderboard();
        updateHeaderStats();
    } catch (error) {
        console.error('Error:', error);
    }
}

function updateHeaderStats() {
    document.getElementById('active-strategies').textContent = leaderboardData.length;
    const totalTrades = leaderboardData.reduce((sum, s) => sum + (s.tradeCount || 0), 0);
    document.getElementById('total-volume').textContent = formatNumber(totalTrades, 0);
}

function renderLeaderboard() {
    const container = document.getElementById('strategy-list');
    const searchTerm = document.getElementById('search-input')?.value?.toLowerCase() || '';

    let filtered = leaderboardData;
    if (searchTerm) {
        filtered = leaderboardData.filter(s =>
            (s.name && s.name.toLowerCase().includes(searchTerm)) ||
            (s.description && s.description.toLowerCase().includes(searchTerm))
        );
    }

    if (filtered.length === 0) {
        container.innerHTML = '<div class="empty-state">暂无策略数据</div>';
        return;
    }

    container.innerHTML = filtered.map((strategy, index) => {
        const rank = index + 1;
        const rankClass = rank <= 3 ? 'top3' : '';
        const pnl = parseFloat(strategy.totalPnl) || 0;
        const roi = parseFloat(strategy.roi) || 0;

        return `
            <div class="strategy-item ${currentStrategy?.apiKey === strategy.apiKey ? 'active' : ''}"
                 onclick="selectStrategy('${strategy.apiKey}')">
                <div class="strategy-header">
                    <div>
                        <span class="strategy-rank ${rankClass}">${rank}</span>
                        <span class="strategy-name">${strategy.name || '未命名'}</span>
                    </div>
                    <span class="strategy-pnl ${getROIColor(pnl)}">
                        ${pnl >= 0 ? '+' : ''}${formatNumber(pnl)}
                    </span>
                </div>
                <div class="strategy-stats">
                    <div class="strategy-stat">
                        <span>ROI</span>
                        <span class="strategy-stat-value ${getROIColor(roi)}">${roi >= 0 ? '+' : ''}${formatNumber(roi)}%</span>
                    </div>
                    <div class="strategy-stat">
                        <span>成交</span>
                        <span class="strategy-stat-value">${strategy.tradeCount || 0}</span>
                    </div>
                    <div class="strategy-stat">
                        <span>初始资金</span>
                        <span class="strategy-stat-value">${formatNumber(strategy.initialBalance, 0)}</span>
                    </div>
                </div>
            </div>
        `;
    }).join('');
}

// Select Strategy
async function selectStrategy(apiKey) {
    currentStrategy = leaderboardData.find(s => s.apiKey === apiKey);
    if (!currentStrategy) return;

    renderLeaderboard();

    const panel = document.getElementById('detail-panel');
    panel.innerHTML = '<div class="loading"><div class="spinner"></div>加载中...</div>';

    try {
        const [statsRes, positionsRes, tradesRes, ordersRes, snapshotsRes, orderbookRes] = await Promise.all([
            fetch(`${API_BASE}/api/dashboard/strategy/${apiKey}`),
            fetch(`${API_BASE}/api/dashboard/strategy/${apiKey}/positions`),
            fetch(`${API_BASE}/api/dashboard/strategy/${apiKey}/trades`),
            fetch(`${API_BASE}/api/dashboard/strategy/${apiKey}/orders`),
            fetch(`${API_BASE}/api/dashboard/strategy/${apiKey}/snapshots?limit=144`),
            fetch(`${API_BASE}/api/dashboard/orderbook/BTCUSDT`)
        ]);

        const stats = await statsRes.json();
        const positions = await positionsRes.json();
        const trades = await tradesRes.json();
        const orders = await ordersRes.json();
        const snapshots = await snapshotsRes.json();
        const orderbook = await orderbookRes.json();

        // Check for API errors
        if (stats.error) {
            panel.innerHTML = `<div class="empty-state"><h3>加载失败</h3><p>${stats.error}</p></div>`;
            return;
        }

        // Ensure arrays exist
        const safePositions = Array.isArray(positions) ? positions : [];
        const safeTrades = Array.isArray(trades) ? trades : [];
        const safeOrders = Array.isArray(orders) ? orders : [];
        const safeSnapshots = Array.isArray(snapshots) ? snapshots : [];
        const safeOrderbook = orderbook || { asks: [], bids: [] };

        renderStrategyDetail(stats, safePositions, safeTrades, safeOrders, safeSnapshots, safeOrderbook);

        // Auto refresh every 5 seconds
        if (refreshInterval) clearInterval(refreshInterval);
        refreshInterval = setInterval(() => refreshDetail(apiKey), 5000);
    } catch (error) {
        console.error('Error:', error);
        panel.innerHTML = '<div class="empty-state"><h3>加载失败</h3><p>无法获取策略详情，请稍后重试</p></div>';
    }
}

async function refreshDetail(apiKey) {
    if (!currentStrategy || currentStrategy.apiKey !== apiKey) return;

    try {
        const [positionsRes, tradesRes, ordersRes, snapshotsRes] = await Promise.all([
            fetch(`${API_BASE}/api/dashboard/strategy/${apiKey}/positions`),
            fetch(`${API_BASE}/api/dashboard/strategy/${apiKey}/trades`),
            fetch(`${API_BASE}/api/dashboard/strategy/${apiKey}/orders`),
            fetch(`${API_BASE}/api/dashboard/strategy/${apiKey}/snapshots?limit=144`)
        ]);

        const positions = await positionsRes.json();
        const trades = await tradesRes.json();
        const orders = await ordersRes.json();
        const snapshots = await snapshotsRes.json();

        // Ensure arrays exist
        const safePositions = Array.isArray(positions) ? positions : [];
        const safeTrades = Array.isArray(trades) ? trades : [];
        const safeOrders = Array.isArray(orders) ? orders : [];
        const safeSnapshots = Array.isArray(snapshots) ? snapshots : [];

        updateDynamicSections(safePositions, safeTrades, safeOrders, safeSnapshots);
    } catch (error) {
        console.error('Refresh error:', error);
    }
}

function updateDynamicSections(positions, trades, orders, snapshots) {
    // Update positions
    const positionsBody = document.getElementById('positions-body');
    if (positionsBody) {
        positionsBody.innerHTML = renderPositionsRows(positions);
    }

    // Update orders
    const ordersList = document.getElementById('orders-list');
    if (ordersList) {
        ordersList.innerHTML = renderOrdersList(orders);
    }

    // Update trades
    const tradesBody = document.getElementById('trades-body');
    if (tradesBody) {
        tradesBody.innerHTML = renderTradesRows(trades.slice(0, 20));
    }

    // Update PnL chart
    if (snapshots.length > 0) {
        drawPnLChart(snapshots);
    }
}

function renderStrategyDetail(stats, positions, trades, orders, snapshots, orderbook) {
    const panel = document.getElementById('detail-panel');

    const available = parseFloat(stats.available) || 0;
    const frozen = parseFloat(stats.frozen) || 0;
    const totalPnl = parseFloat(stats.totalPnl) || 0;
    const roi = parseFloat(stats.roi) || 0;
    const initialBalance = parseFloat(stats.initialBalance) || 0;
    const totalEquity = available + frozen;
    const marginUsage = initialBalance > 0 ? (frozen / initialBalance * 100) : 0;

    let riskLevel = 'low', riskText = '低风险';
    if (marginUsage > 50) { riskLevel = 'high'; riskText = '高风险'; }
    else if (marginUsage > 20) { riskLevel = 'medium'; riskText = '中风险'; }

    panel.innerHTML = `
        <div class="top-stats">
            <div class="top-stat-card">
                <div class="top-stat-label">累计收益</div>
                <div class="top-stat-value ${getROIColor(totalPnl)}">${totalPnl >= 0 ? '+' : ''}${formatNumber(totalPnl)} USDT</div>
                <div class="top-stat-change ${getROIColor(roi)}">${roi >= 0 ? '+' : ''}${formatNumber(roi)}% ROI</div>
            </div>
            <div class="top-stat-card">
                <div class="top-stat-label">当前权益</div>
                <div class="top-stat-value">${formatNumber(totalEquity)} USDT</div>
                <div class="top-stat-change" style="color: var(--text-muted);">初始: ${formatNumber(initialBalance)}</div>
            </div>
            <div class="top-stat-card">
                <div class="top-stat-label">可用余额</div>
                <div class="top-stat-value">${formatNumber(available)} USDT</div>
                <div class="top-stat-change" style="color: var(--text-muted);">冻结: ${formatNumber(frozen)}</div>
            </div>
            <div class="top-stat-card">
                <div class="top-stat-label">持仓数量</div>
                <div class="top-stat-value">${positions.length}</div>
                <div class="top-stat-change" style="color: var(--text-muted);">未成交: ${orders.length}</div>
            </div>
            <div class="top-stat-card">
                <div class="top-stat-label">风险等级</div>
                <div class="risk-meter" style="margin-top: 8px;">
                    <div class="risk-bar">
                        <div class="risk-fill ${riskLevel}" style="width: ${Math.min(marginUsage, 100)}%"></div>
                    </div>
                    <span class="risk-text ${riskLevel}">${riskText}</span>
                </div>
                <div class="top-stat-change" style="color: var(--text-muted); margin-top: 4px;">保证金占用: ${formatNumber(marginUsage)}%</div>
            </div>
        </div>

        <div class="content-grid">
            <div class="left-column">
                <div class="card">
                    <div class="card-header">
                        <span class="card-title">收益曲线 (24H)</span>
                    </div>
                    <div class="card-body">
                        <div class="chart-container" id="pnl-chart">
                            ${snapshots.length > 0 ? drawPnLChart(snapshots) : '<div class="empty-state">暂无数据</div>'}
                        </div>
                    </div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <span class="card-title">当前持仓</span>
                        <span class="card-badge">${positions.length}</span>
                    </div>
                    <div class="card-body" style="padding: 0;">
                        <table class="table">
                            <thead>
                                <tr>
                                    <th>交易对</th>
                                    <th>方向</th>
                                    <th>开仓价格</th>
                                    <th>数量</th>
                                    <th>杠杆</th>
                                    <th>保证金</th>
                                    <th>未实现盈亏</th>
                                </tr>
                            </thead>
                            <tbody id="positions-body">
                                ${renderPositionsRows(positions)}
                            </tbody>
                        </table>
                    </div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <span class="card-title">历史成交</span>
                        <span class="card-badge">${trades.length}</span>
                    </div>
                    <div class="card-body" style="padding: 0;">
                        <table class="table">
                            <thead>
                                <tr>
                                    <th>时间</th>
                                    <th>交易对</th>
                                    <th>方向</th>
                                    <th>价格</th>
                                    <th>数量</th>
                                    <th>成交额</th>
                                    <th>手续费</th>
                                </tr>
                            </thead>
                            <tbody id="trades-body">
                                ${renderTradesRows(trades.slice(0, 20))}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            <div class="right-column">
                <div class="card">
                    <div class="card-header">
                        <span class="card-title">订单簿 BTCUSDT</span>
                    </div>
                    <div class="card-body" style="padding: 0;">
                        <div class="orderbook">
                            <div class="orderbook-header" style="padding: 8px 12px;">
                                <span>价格</span>
                                <span>数量</span>
                                <span>累计</span>
                            </div>
                            <div class="orderbook-asks" style="max-height: 150px; overflow: hidden;">
                                ${renderOrderbookAsks(orderbook.asks.slice().reverse())}
                            </div>
                            <div style="padding: 8px 12px; background: var(--bg-tertiary); text-align: center; font-weight: 600; color: var(--text-primary);">
                                ${orderbook.asks.length > 0 ? formatPrice(orderbook.asks[0].price) : '-'}
                            </div>
                            <div class="orderbook-bids" style="max-height: 150px; overflow: hidden;">
                                ${renderOrderbookBids(orderbook.bids)}
                            </div>
                        </div>
                    </div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <span class="card-title">当前委托</span>
                        <span class="card-badge">${orders.length}</span>
                    </div>
                    <div class="card-body" style="padding: 0;" id="orders-list">
                        ${renderOrdersList(orders)}
                    </div>
                </div>
            </div>
        </div>
    `;

    if (snapshots.length > 0) {
        drawPnLChart(snapshots);
    }
}

function renderPositionsRows(positions) {
    if (positions.length === 0) {
        return '<tr><td colspan="7" class="empty-state">暂无持仓</td></tr>';
    }

    return positions.map(p => {
        const pnl = parseFloat(p.unrealizedPnl) || 0;
        return `
            <tr>
                <td class="symbol">${p.symbol}</td>
                <td class="${p.side === 'LONG' ? 'side-long' : 'side-short'}">${p.side}</td>
                <td>${formatPrice(p.entryPrice)}</td>
                <td>${formatNumber(p.size, 4)}</td>
                <td>${p.leverage}x</td>
                <td>${formatNumber(p.margin)} USDT</td>
                <td class="${pnl >= 0 ? 'pnl-positive' : 'pnl-negative'}">${pnl >= 0 ? '+' : ''}${formatNumber(pnl)}</td>
            </tr>
        `;
    }).join('');
}

function renderTradesRows(trades) {
    if (trades.length === 0) {
        return '<tr><td colspan="7" class="empty-state">暂无成交记录</td></tr>';
    }

    return trades.map(t => `
        <tr>
            <td>${formatTime(t.time || t.timestamp)}</td>
            <td class="symbol">${t.symbol}</td>
            <td class="${t.side === 'BUY' ? 'side-long' : 'side-short'}">${t.side}</td>
            <td>${formatPrice(t.price)}</td>
            <td>${formatNumber(t.qty || t.quantity, 4)}</td>
            <td>${formatNumber(t.quoteQty)} USDT</td>
            <td>${formatNumber(t.fee, 4)}</td>
        </tr>
    `).join('');
}

function renderOrdersList(orders) {
    if (orders.length === 0) {
        return '<div class="empty-state" style="padding: 24px;">暂无未成交订单</div>';
    }

    return orders.map(o => `
        <div class="order-item">
            <div class="order-info">
                <div class="order-symbol">${o.symbol} <span style="color: var(--text-muted);">${o.side}</span></div>
                <div class="order-details">限价 ${formatPrice(o.price)} | ${o.leverage}x杠杆</div>
            </div>
            <div class="order-price">
                <div class="order-amount">${formatNumber(o.quantity, 4)}</div>
                <div class="order-type">${o.status}</div>
            </div>
        </div>
    `).join('');
}

function renderOrderbookAsks(asks) {
    return asks.map(a => {
        const qty = parseFloat(a.quantity);
        const width = Math.min((qty / 10) * 100, 100);
        return `
            <div class="orderbook-row ask">
                <div class="orderbook-bar ask" style="width: ${width}%"></div>
                <span>${formatPrice(a.price)}</span>
                <span>${formatNumber(qty, 4)}</span>
                <span style="color: var(--text-muted);">-</span>
            </div>
        `;
    }).join('');
}

function renderOrderbookBids(bids) {
    return bids.map(b => {
        const qty = parseFloat(b.quantity);
        const width = Math.min((qty / 10) * 100, 100);
        return `
            <div class="orderbook-row bid">
                <div class="orderbook-bar bid" style="width: ${width}%"></div>
                <span>${formatPrice(b.price)}</span>
                <span>${formatNumber(qty, 4)}</span>
                <span style="color: var(--text-muted);">-</span>
            </div>
        `;
    }).join('');
}

function drawPnLChart(snapshots) {
    const container = document.getElementById('pnl-chart');
    if (!container || snapshots.length === 0) return;

    const sorted = snapshots.slice().sort((a, b) => new Date(a.snapshotAt) - new Date(b.snapshotAt));
    const values = sorted.map(s => parseFloat(s.totalPnl) || 0);
    const min = Math.min(...values);
    const max = Math.max(...values);
    const range = max - min || 1;

    const width = container.clientWidth || 800;
    const height = 200;
    const padding = 20;

    const points = values.map((v, i) => {
        const x = padding + (i / (values.length - 1 || 1)) * (width - padding * 2);
        const y = height - padding - ((v - min) / range) * (height - padding * 2);
        return `${x},${y}`;
    }).join(' ');

    const color = values[values.length - 1] >= values[0] ? 'var(--color-long)' : 'var(--color-short)';

    container.innerHTML = `
        <svg class="chart-svg" viewBox="0 0 ${width} ${height}" preserveAspectRatio="none">
            <defs>
                <linearGradient id="pnlGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stop-color="${color}" stop-opacity="0.3"/>
                    <stop offset="100%" stop-color="${color}" stop-opacity="0"/>
                </linearGradient>
            </defs>
            <polygon points="${padding},${height-padding} ${points} ${width-padding},${height-padding}" fill="url(#pnlGradient)"/>
            <polyline points="${points}" fill="none" stroke="${color}" stroke-width="2"/>
        </svg>
    `;
}

// Search functionality
document.addEventListener('DOMContentLoaded', () => {
    loadLeaderboard();
    setInterval(loadLeaderboard, 5000);

    document.getElementById('search-input')?.addEventListener('input', renderLeaderboard);
});

window.addEventListener('beforeunload', () => {
    if (refreshInterval) clearInterval(refreshInterval);
});
