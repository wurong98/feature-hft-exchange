async function loadLeaderboard() {
    // TODO: 实现从后端获取数据
    const mockData = [
        { rank: 1, name: 'MakerBot v2', desc: '联系: maker@example.com', profit: 25.5, trades: 150, winRate: 65 },
        { rank: 2, name: 'Grid Hunter', desc: '高频网格策略', profit: 18.2, trades: 320, winRate: 58 },
    ];

    const tbody = document.querySelector('#leaderboard tbody');
    tbody.innerHTML = mockData.map(s => `
        <tr>
            <td class="rank">#${s.rank}</td>
            <td>
                <div class="strategy-name">${s.name}</div>
                <div class="strategy-desc">${s.desc}</div>
            </td>
            <td class="${s.profit >= 0 ? 'profit' : 'loss'}">${s.profit >= 0 ? '+' : ''}${s.profit}%</td>
            <td>${s.trades}</td>
            <td>${s.winRate}%</td>
        </tr>
    `).join('');
}

loadLeaderboard();
setInterval(loadLeaderboard, 5000);
