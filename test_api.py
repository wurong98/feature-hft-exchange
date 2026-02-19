#!/usr/bin/env python3
"""
HFT Simulated Exchange API 测试脚本
直接调用 REST API 测试挂单、查询、取消等功能
"""

import argparse
import json
import sys
import time
from typing import Optional, Dict, List

import requests


class HFTExchangeTester:
    """HFT 模拟交易所测试器"""

    def __init__(self, api_key: str, base_url: str = "http://localhost:8080"):
        self.api_key = api_key
        self.base_url = base_url.rstrip('/')
        self.headers = {
            'X-MBX-APIKEY': api_key,
            'Content-Type': 'application/json'
        }

        print(f"[*] 连接到 HFT 模拟交易所: {base_url}")
        print(f"[*] API Key: {api_key[:8]}...{api_key[-4:]}")

    def _request(self, method: str, endpoint: str, data: Optional[Dict] = None, params: Optional[Dict] = None) -> Dict:
        """发送 HTTP 请求"""
        url = f"{self.base_url}{endpoint}"

        try:
            if method == 'GET':
                response = requests.get(url, headers=self.headers, params=params)
            elif method == 'POST':
                response = requests.post(url, headers=self.headers, json=data)
            elif method == 'DELETE':
                response = requests.delete(url, headers=self.headers, params=params)
            else:
                raise ValueError(f"不支持的 HTTP 方法: {method}")

            response.raise_for_status()
            return response.json()

        except requests.exceptions.RequestException as e:
            print(f"    [-] 请求失败: {e}")
            return {'error': str(e)}

    def test_exchange_info(self) -> bool:
        """测试交易所信息"""
        print("\n[+] 测试交易所信息...")
        result = self._request('GET', '/api/v3/exchangeInfo')

        if 'error' in result:
            print(f"    [-] 失败: {result['error']}")
            return False

        symbols = result.get('symbols', [])
        print(f"    [✓] 支持的交易对: {len(symbols)} 个")
        for sym in symbols[:5]:
            print(f"        - {sym['symbol']}: {sym['baseAsset']}/{sym['quoteAsset']}")

        return True

    def test_account(self) -> bool:
        """测试账户查询"""
        print("\n[+] 测试账户查询...")
        result = self._request('GET', '/api/v3/account')

        if 'error' in result or result.get('code'):
            print(f"    [-] 失败: {result.get('msg', result.get('error', 'Unknown'))}")
            return False

        balances = result.get('balances') or []
        positions = result.get('positions') or []

        print(f"    [✓] 账户查询成功")
        print(f"        账户类型: {result.get('accountType', 'N/A')}")

        for bal in balances:
            if float(bal.get('free', 0)) > 0 or float(bal.get('locked', 0)) > 0:
                print(f"        {bal.get('asset', 'N/A')}: 可用={bal.get('free', 0)}, 冻结={bal.get('locked', 0)}")

        print(f"        当前持仓: {len(positions)} 个")
        for pos in positions:
            print(f"          - {pos.get('symbol', 'N/A')}: {pos.get('side', 'N/A')} "
                  f"{pos.get('size', 0)} @ {pos.get('entryPrice', 0)}")

        return True

    def test_create_order(self, symbol: str, side: str, price: float,
                         quantity: float, leverage: int = 10) -> Optional[str]:
        """测试创建限价单"""
        print(f"\n[+] 测试创建订单...")
        print(f"    交易对: {symbol}")
        print(f"    方向: {side}")
        print(f"    价格: {price}")
        print(f"    数量: {quantity}")
        print(f"    杠杆: {leverage}x")

        data = {
            'symbol': symbol,
            'side': side,
            'type': 'LIMIT',
            'price': str(price),
            'quantity': str(quantity),
            'leverage': leverage
        }

        result = self._request('POST', '/api/v3/order', data=data)

        if result.get('code') and result['code'] < 0:
            print(f"    [-] 失败: {result.get('msg', 'Unknown error')}")
            return None

        if 'error' in result:
            print(f"    [-] 失败: {result['error']}")
            return None

        print(f"    [✓] 订单创建成功")
        print(f"        订单ID: {result.get('orderId')}")
        print(f"        状态: {result.get('status')}")
        print(f"        价格: {result.get('price')}")
        print(f"        数量: {result.get('origQty')}")

        return str(result.get('orderId'))

    def test_get_open_orders(self, symbol: str) -> List[Dict]:
        """测试查询未成交订单"""
        print(f"\n[+] 测试查询未成交订单 ({symbol})...")

        params = {'symbol': symbol} if symbol else {}
        result = self._request('GET', '/api/v3/openOrders', params=params)

        if isinstance(result, dict) and 'error' in result:
            print(f"    [-] 失败: {result['error']}")
            return []

        if not isinstance(result, list):
            print(f"    [-] 响应格式错误")
            return []

        print(f"    [✓] 未成交订单: {len(result)} 个")
        for order in result:
            print(f"      - ID: {order.get('orderId')}, {order.get('side')} "
                  f"{order.get('origQty')} @ {order.get('price')}, 状态: {order.get('status')}")

        return result

    def test_cancel_order(self, order_id: str, symbol: str) -> bool:
        """测试取消订单"""
        print(f"\n[+] 测试取消订单...")
        print(f"    订单ID: {order_id}")

        params = {'orderId': order_id, 'symbol': symbol}
        result = self._request('DELETE', '/api/v3/order', params=params)

        if result.get('code') and result['code'] < 0:
            print(f"    [-] 失败: {result.get('msg', 'Unknown error')}")
            return False

        if 'error' in result:
            print(f"    [-] 失败: {result['error']}")
            return False

        print(f"    [✓] 取消成功")
        print(f"        状态: {result.get('status', 'CANCELLED')}")
        return True

    def test_get_trades(self, symbol: str) -> List[Dict]:
        """测试查询成交记录"""
        print(f"\n[+] 测试查询成交记录 ({symbol})...")

        params = {'symbol': symbol} if symbol else {}
        result = self._request('GET', '/api/v3/myTrades', params=params)

        if isinstance(result, dict) and 'error' in result:
            print(f"    [-] 失败: {result['error']}")
            return []

        if not isinstance(result, list):
            print(f"    [✓] 历史成交: 0 笔")
            return []

        print(f"    [✓] 历史成交: {len(result)} 笔")
        for trade in result[:5]:
            print(f"      - {trade.get('time', 'N/A')}: {trade.get('side')} "
                  f"{trade.get('qty')} @ {trade.get('price')}, 手续费: {trade.get('fee', 0)}")

        if len(result) > 5:
            print(f"      ... 还有 {len(result) - 5} 笔")

        return result

    def test_get_orderbook(self, symbol: str) -> bool:
        """测试查询订单簿"""
        print(f"\n[+] 测试查询订单簿 ({symbol})...")

        result = self._request('GET', f'/api/dashboard/orderbook/{symbol}')

        if 'error' in result:
            print(f"    [-] 失败: {result['error']}")
            return False

        bids = result.get('bids', [])
        asks = result.get('asks', [])

        print(f"    [✓] 订单簿查询成功")
        print(f"        买盘: {len(bids)} 档")
        print(f"        卖盘: {len(asks)} 档")

        if asks:
            print(f"        最佳卖价: {asks[0]['price']}")
        if bids:
            print(f"        最佳买价: {bids[0]['price']}")

        return True

    def run_full_test(self, symbol: str, test_cancel: bool = True) -> bool:
        """运行完整测试流程"""
        print("=" * 60)
        print("HFT 模拟交易所 API 完整测试")
        print("=" * 60)

        # 1. 交易所信息
        self.test_exchange_info()

        # 2. 账户查询
        self.test_account()

        # 3. 订单簿
        self.test_get_orderbook(symbol)

        # 4. 创建买单
        buy_price = 65000
        buy_order_id = self.test_create_order(
            symbol=symbol,
            side='BUY',
            price=buy_price,
            quantity=0.01,
            leverage=10
        )

        if not buy_order_id:
            print("\n[-] 买单创建失败，继续测试...")
        else:
            time.sleep(0.5)

        # 5. 创建卖单
        sell_price = 75000
        sell_order_id = self.test_create_order(
            symbol=symbol,
            side='SELL',
            price=sell_price,
            quantity=0.01,
            leverage=10
        )

        if not sell_order_id:
            print("\n[-] 卖单创建失败")
        else:
            time.sleep(0.5)

        # 6. 查询未成交订单
        open_orders = self.test_get_open_orders(symbol)

        # 7. 查询成交记录
        self.test_get_trades(symbol)

        # 8. 取消订单
        if test_cancel:
            if buy_order_id:
                time.sleep(0.5)
                self.test_cancel_order(buy_order_id, symbol)

            if sell_order_id:
                time.sleep(0.5)
                self.test_cancel_order(sell_order_id, symbol)

            # 9. 再次查询未成交订单
            time.sleep(0.5)
            self.test_get_open_orders(symbol)

        print("\n" + "=" * 60)
        print("测试完成")
        print("=" * 60)

        return True


def main():
    parser = argparse.ArgumentParser(
        description='HFT 模拟交易所 API 测试工具',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
使用示例:
  # 完整测试（需要 API Key）
  python test_api.py --api-key YOUR_API_KEY

  # 指定自定义 URL
  python test_api.py --api-key YOUR_API_KEY --url http://192.168.1.100:8080

  # 测试指定交易对
  python test_api.py --api-key YOUR_API_KEY --symbol ETHUSDT

  # 只测试账户查询
  python test_api.py --api-key YOUR_API_KEY --action account

  # 创建测试订单（不取消）
  python test_api.py --api-key YOUR_API_KEY --action create --no-cancel

  # 查询订单簿
  python test_api.py --api-key YOUR_API_KEY --action orderbook
        """
    )

    parser.add_argument(
        '--api-key',
        required=True,
        help='API Key (必需)'
    )

    parser.add_argument(
        '--url',
        default='http://localhost:8080',
        help='交易所基础 URL (默认: http://localhost:8080)'
    )

    parser.add_argument(
        '--symbol',
        default='BTCUSDT',
        help='交易对 (默认: BTCUSDT)'
    )

    parser.add_argument(
        '--action',
        choices=['full', 'account', 'create', 'query', 'cancel', 'orderbook'],
        default='full',
        help='测试动作 (默认: full)'
    )

    parser.add_argument(
        '--no-cancel',
        action='store_true',
        help='创建订单后不自动取消（用于手动测试）'
    )

    parser.add_argument(
        '--price',
        type=float,
        default=65000,
        help='挂单价格 (默认: 65000)'
    )

    parser.add_argument(
        '--quantity',
        type=float,
        default=0.01,
        help='挂单数量 (默认: 0.01)'
    )

    parser.add_argument(
        '--leverage',
        type=int,
        default=10,
        help='杠杆倍数 (默认: 10)'
    )

    args = parser.parse_args()

    # 创建测试器
    tester = HFTExchangeTester(args.api_key, args.url)

    # 执行测试
    if args.action == 'full':
        tester.run_full_test(args.symbol, test_cancel=not args.no_cancel)

    elif args.action == 'account':
        tester.test_account()

    elif args.action == 'create':
        order_id = tester.test_create_order(
            args.symbol,
            'BUY',
            args.price,
            args.quantity,
            args.leverage
        )
        if order_id and not args.no_cancel:
            time.sleep(1)
            tester.test_cancel_order(order_id, args.symbol)

    elif args.action == 'query':
        tester.test_get_open_orders(args.symbol)
        tester.test_get_trades(args.symbol)

    elif args.action == 'cancel':
        orders = tester.test_get_open_orders(args.symbol)
        if orders:
            for order in orders:
                tester.test_cancel_order(str(order['orderId']), args.symbol)
                time.sleep(0.3)
        else:
            print("没有可取消的订单")

    elif args.action == 'orderbook':
        tester.test_get_orderbook(args.symbol)


if __name__ == '__main__':
    main()
