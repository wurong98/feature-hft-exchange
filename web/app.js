// 状态管理
let currentStrategy = null;
let leaderboardData = [];

// API 基础 URL
const API_BASE = '';

// 格式化数字
function formatNumber(num, decimals = 2) {
    if (num === undefined || num === null || num === '') return '-';
    const n = parseFloat(num);
    if (isNaN(n)) return '-';
    return n.toFixed(decimals);
}

// 格式化收益率
function formatROI(roi) {
    const sign = roi >= 0 ? '+' : '';
    const className = roi >= 0 ? 'profit' : 'loss';
    return `<span class="${className}">${sign}${formatNumber(roi)}%</span>`;
}

// 加载排行榜
async function loadLeaderboard() {
    try {
        const response = await fetch(`${API_BASE}/api/dashboard/leaderboard`);
        if (!response.ok) throw new Error('Failed to load leaderboard');

        leaderboardData = await response.json();
        renderLeaderboard();

        document.getElementById('connection-status').textContent = '已连接';
        document.getElementById('connection-status').style.color = '#3fb950';
    } catch (error) {
        console.error('Error loading leaderboard:', error);
        document.getElementById('connection-status').textContent = '连接失败';
        document.getElementById('connection-status').style.color = '#f85149';
    }
}

// 渲染排行榜
function renderLeaderboard() {
    const container = document.getElementById('strategy-list');

    if (leaderboardData.length === 0) {
        container.innerHTML = '<div class="empty-state">暂无策略数据</div>';
        return;
    }

    container.innerHTML = leaderboardData.map((strategy, index) => {
        const totalPnl = parseFloat(strategy.totalPnl) || 0;
        const roi = parseFloat(strategy.roi) || 0;
        return `
        <div class="strategy-item ${currentStrategy?.apiKey === strategy.apiKey ? 'active' : ''}"
             onclick="selectStrategy('${strategy.apiKey}')">
            <div class="strategy-header">
                <span class="strategy-name">${strategy.name || '未命名策略'}</span>
                <span class="strategy-rank">#${index + 1}</span>
            </div>
            <div class="strategy-desc">${strategy.description || '暂无描述'}</div>
            <div class="strategy-stats">
                <div>
                    <span class="stat-value ${totalPnl >= 0 ? 'profit' : 'loss'}">
                        ${totalPnl >= 0 ? '+' : ''}${formatNumber(totalPnl)}
                    </span>
                    <span class="stat-label">USDT</span>
                </div>
                <div>
                    <span class="stat-value">${strategy.tradeCount || 0}</span>
                    <span class="stat-label">成交</span>
                </div>
                <div>
                    <span class="stat-value ${roi >= 0 ? 'profit' : 'loss'}">
                        ${roi >= 0 ? '+' : ''}${formatNumber(roi)}%
                    </span>
                    <span class="stat-label">ROI</span>
                </div>
            </div>
        </div>
    `}).join('');
}

// 选择策略
async function selectStrategy(apiKey) {
    currentStrategy = leaderboardData.find(s => s.apiKey === apiKey);
    renderLeaderboard(); // 重新渲染以更新 active 状态

    const detailPanel = document.getElementById('detail-panel');
    detailPanel.innerHTML = `
        <div class="loading">
            <div class="spinner"></div>
            加载策略详情...
        </div>
    `;

    try {
        // 并行加载数据
        const [stats, positions, trades] = await Promise.all([
            fetch(`${API_BASE}/api/dashboard/strategy/${apiKey}`).then(r => r.json()),
            fetch(`${API_BASE}/api/dashboard/strategy/${apiKey}/positions`).then(r => r.json()),
            fetch(`${API_BASE}/api/dashboard/strategy/${apiKey}/trades`).then(r => r.json())
        ]);

        renderStrategyDetail(stats, positions, trades);
    } catch (error) {
        console.error('Error loading strategy detail:', error);
        detailPanel.innerHTML = `
            <div class="detail-empty">
                <h2>加载失败</h2>
                <p>无法获取策略详情，请稍后重试</p>
            </div>
        `;
    }
}

// 渲染策略详情
function renderStrategyDetail(stats, positions, trades) {
    const detailPanel = document.getElementById('detail-panel');

    const available = parseFloat(stats.available) || 0;
    const frozen = parseFloat(stats.frozen) || 0;
    const totalPnl = parseFloat(stats.totalPnl) || 0;
    const roi = parseFloat(stats.roi) || 0;
    const initialBalance = parseFloat(stats.initialBalance) || 0;

    const totalEquity = available + frozen;
    const marginUsage = initialBalance > 0 ? (frozen / initialBalance * 100) : 0;

    // 计算风险等级
    let riskClass = 'risk-low';
    let riskText = '低风险';
    if (marginUsage > 50) {
        riskClass = 'risk-high';
        riskText = '高风险';
    } else if (marginUsage > 20) {
        riskClass = 'risk-medium';
        riskText = '中风险';
    }

    detailPanel.innerHTML = `
        <div class="detail-header">
            <h2>${stats.name || '未命名策略'}</h2>
            <span class="api-key">${stats.apiKey}</span>
        </div>

        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-card-title">累计收益 (USDT)</div>
                <div class="stat-card-value ${totalPnl >= 0 ? 'profit' : 'loss'}">
                    ${totalPnl >= 0 ? '+' : ''}${formatNumber(totalPnl)}
                </div>
                <div class="stat-card-sub">初始资金: ${formatNumber(initialBalance)} USDT</div>
            </div>

            <div class="stat-card">
                <div class="stat-card-title">收益率 (ROI)</div>
                <div class="stat-card-value ${roi >= 0 ? 'profit' : 'loss'}">
                    ${roi >= 0 ? '+' : ''}${formatNumber(roi)}%
                </div>
                <div class="stat-card-sub">总交易: ${stats.tradeCount || 0} 笔</div>
            </div>

            <div class="stat-card">
                <div class="stat-card-title">当前权益</div>
                <div class="stat-card-value">${formatNumber(totalEquity)}</div>
                <div class="stat-card-sub">可用: ${formatNumber(available)} / 冻结: ${formatNumber(frozen)}</div>
            </div>

            <div class="stat-card">
                <div class="stat-card-title">风险等级</div>
                <div class="stat-card-value ${riskClass}">${riskText}</div>
                <div class="stat-card-sub">保证金占用: ${formatNumber(marginUsage)}%</div>
            </div>
        </div>

        <div class="section">
            <div class="section-header">当前持仓 (${positions.length})</div>
            ${positions.length > 0 ? `
                <table>
                    <thead>
                        <tr>
                            <th>交易对</th>
                            <th>方向</th>
                            <th>开仓价格</th>
                            <th>持仓数量</th>
                            <th>杠杆</th>
                            <th>保证金</th>
                            <th>未实现盈亏</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${positions.map(p => {
                            const unrealizedPnl = parseFloat(p.unrealizedPnl) || 0;
                            return `
                            <tr>
                                <td>${p.symbol}</td>
                                <td class="${p.side === 'LONG' ? 'side-long' : 'side-short'}">${p.side}</td>
                                <td>${formatNumber(p.entryPrice)}</td>
                                <td>${formatNumber(p.size)}</td>
                                <td>${p.leverage}x</td>
                                <td>${formatNumber(p.margin)} USDT</td>
                                <td class="${unrealizedPnl >= 0 ? 'profit' : 'loss'}">
                                    ${unrealizedPnl >= 0 ? '+' : ''}${formatNumber(unrealizedPnl)}
                                </td>
                            </tr>
                        `}).join('')}
                    </tbody>
                </table>
            ` : '<div class="empty-state">暂无持仓</div>'}
        </div>

        <div class="section">
            <div class="section-header">历史成交 (${trades.length})</div>
            ${trades.length > 0 ? `
                <table>
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
                    <tbody>
                        ${trades.slice(0, 50).map(t => `
                            <tr>
                                <td>${formatTime(t.time)}</td>
                                <td>${t.symbol}</td>
                                <td class="${t.side === 'BUY' ? 'side-long' : 'side-short'}">${t.side}</td>
                                <td>${formatNumber(t.price)}</td>
                                <td>${formatNumber(t.qty || t.quantity)}</td>
                                <td>${formatNumber(t.quoteQty)} USDT</td>
                                <td>${formatNumber(t.fee)}</td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
                ${trades.length > 50 ? `<div class="empty-state">还有 ${trades.length - 50} 条记录...</div>` : ''}
            ` : '<div class="empty-state">暂无成交记录</div>'}
        </div>
    `;
}

// 格式化时间
function formatTime(timeStr) {
    if (!timeStr) return '-';
    const date = new Date(timeStr);
    return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

// 初始化
loadLeaderboard();
setInterval(loadLeaderboard, 5000); // 每5秒刷新排行榜
