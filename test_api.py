#!/usr/bin/env python3
"""
HFT Simulated Exchange API 测试脚本
使用 CCXT 库测试挂单、查询、取消等功能
"""

import argparse
import sys
import time

import ccxt


class HFTExchangeTester:
    """HFT 模拟交易所测试器"""

    def __init__(self, api_key, base_url="http://localhost:8080"):
        self.api_key = api_key
        self.base_url = base_url

        # 初始化 ccxt binance 配置
        self.exchange = ccxt.binance({
            'apiKey': api_key,
            'secret': 'dummy-secret',  # 模拟交易所不需要验证签名
            'urls': {
                'api': {
                    'public': f'{base_url}/api/v3',
                    'private': f'{base_url}/api/v3',
                }
            },
            'options': {
                'defaultType': 'future',  # 使用合约模式
            },
            'enableRateLimit': True,
        })

        print(f"[*] 连接到 HFT 模拟交易所: {base_url}")
        print(f"[*] API Key: {api_key[:8]}...{api_key[-4:]}")

    def test_account(self):
        """测试账户查询"""
        print("\n[+] 测试账户查询...")
        try:
            balance = self.exchange.fetch_balance()
            print(f"    账户类型: {balance.get('accountType', 'N/A')}")
            print(f"    USDT 可用: {balance['USDT']['free']}")
            print(f"    USDT 冻结: {balance['USDT']['used']}")

            # 获取持仓
            positions = self.exchange.fetch_positions()
            print(f"    当前持仓: {len(positions)} 个")
            for pos in positions:
                if pos['contracts'] > 0:
                    print(f"      - {pos['symbol']}: {pos['side']} {pos['contracts']} @ {pos['entryPrice']}")

            return True
        except Exception as e:
            print(f"    [-] 失败: {e}")
            return False

    def test_create_order(self, symbol, side, price, quantity, leverage=10):
        """测试创建限价单"""
        print(f"\n[+] 测试创建订单...")
        print(f"    交易对: {symbol}")
        print(f"    方向: {side}")
        print(f"    价格: {price}")
        print(f"    数量: {quantity}")
        print(f"    杠杆: {leverage}x")

        try:
            order = self.exchange.create_limit_buy_order(
                symbol=symbol,
                amount=quantity,
                price=price,
                params={'leverage': leverage} if side == 'BUY' else {}
            )

            if side == 'SELL':
                # 创建卖单
                order = self.exchange.create_limit_sell_order(
                    symbol=symbol,
                    amount=quantity,
                    price=price,
                    params={'leverage': leverage}
                )

            print(f"    [✓] 订单创建成功")
            print(f"        订单ID: {order['id']}")
            print(f"        状态: {order['status']}")
            print(f"        价格: {order['price']}")
            print(f"        数量: {order['amount']}")

            return order['id']
        except Exception as e:
            print(f"    [-] 失败: {e}")
            return None

    def test_get_open_orders(self, symbol):
        """测试查询未成交订单"""
        print(f"\n[+] 测试查询未成交订单 ({symbol})...")
        try:
            orders = self.exchange.fetch_open_orders(symbol)
            print(f"    未成交订单: {len(orders)} 个")

            for order in orders:
                print(f"      - ID: {order['id']}, {order['side']} {order['amount']} @ {order['price']}, 状态: {order['status']}")

            return orders
        except Exception as e:
            print(f"    [-] 失败: {e}")
            return []

    def test_cancel_order(self, order_id, symbol):
        """测试取消订单"""
        print(f"\n[+] 测试取消订单...")
        print(f"    订单ID: {order_id}")

        try:
            result = self.exchange.cancel_order(order_id, symbol)
            print(f"    [✓] 取消成功")
            print(f"        状态: {result.get('status', 'CANCELLED')}")
            return True
        except Exception as e:
            print(f"    [-] 失败: {e}")
            return False

    def test_get_trades(self, symbol):
        """测试查询成交记录"""
        print(f"\n[+] 测试查询成交记录 ({symbol})...")
        try:
            trades = self.exchange.fetch_my_trades(symbol)
            print(f"    历史成交: {len(trades)} 笔")

            for trade in trades[:5]:  # 只显示前5笔
                print(f"      - {trade['datetime']}: {trade['side']} {trade['amount']} @ {trade['price']}, 手续费: {trade['fee']['cost']}")

            if len(trades) > 5:
                print(f"      ... 还有 {len(trades) - 5} 笔")

            return trades
        except Exception as e:
            print(f"    [-] 失败: {e}")
            return []

    def test_exchange_info(self):
        """测试交易所信息"""
        print("\n[+] 测试交易所信息...")
        try:
            markets = self.exchange.load_markets()
            print(f"    支持的交易对: {len(markets)} 个")

            for symbol in list(markets.keys())[:5]:
                market = markets[symbol]
                print(f"      - {symbol}: {market['type']}, 杠杆 {market['limits']['leverage']}")

            return True
        except Exception as e:
            print(f"    [-] 失败: {e}")
            return False

    def run_full_test(self, symbol, test_cancel=True):
        """运行完整测试流程"""
        print("=" * 60)
        print("HFT 模拟交易所 API 完整测试")
        print("=" * 60)

        # 1. 交易所信息
        self.test_exchange_info()

        # 2. 账户查询
        self.test_account()

        # 3. 创建买单
        buy_price = 65000  # 低于市价，不会立即成交
        buy_order_id = self.test_create_order(
            symbol=symbol,
            side='BUY',
            price=buy_price,
            quantity=0.01,
            leverage=10
        )

        if not buy_order_id:
            print("\n[-] 买单创建失败，终止测试")
            return False

        time.sleep(0.5)

        # 4. 创建卖单
        sell_price = 75000  # 高于市价，不会立即成交
        sell_order_id = self.test_create_order(
            symbol=symbol,
            side='SELL',
            price=sell_price,
            quantity=0.01,
            leverage=10
        )

        if not sell_order_id:
            print("\n[-] 卖单创建失败")

        time.sleep(0.5)

        # 5. 查询未成交订单
        open_orders = self.test_get_open_orders(symbol)

        # 6. 查询成交记录
        self.test_get_trades(symbol)

        # 7. 取消订单
        if test_cancel and buy_order_id:
            time.sleep(0.5)
            self.test_cancel_order(buy_order_id, symbol)

            if sell_order_id:
                time.sleep(0.5)
                self.test_cancel_order(sell_order_id, symbol)

        # 8. 再次查询未成交订单
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
        choices=['full', 'account', 'create', 'query', 'cancel'],
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
                tester.test_cancel_order(order['id'], args.symbol)
                time.sleep(0.3)
        else:
            print("没有可取消的订单")


if __name__ == '__main__':
    main()
