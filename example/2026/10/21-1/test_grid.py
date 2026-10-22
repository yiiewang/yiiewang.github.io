"""grid.py 的单元测试：档位推进、配对成交、基准更新、自我保护。

运行：python3 -m unittest test_grid -v
"""

import unittest

from grid import BUY_LEVELS, SELL_LEVELS, GridEngine


class LevelTest(unittest.TestCase):
    def test_fibonacci_levels(self):
        # 累计跌幅序列与博客表格一致
        self.assertEqual(
            tuple(round(x, 3) for x in BUY_LEVELS),
            (0.010, 0.020, 0.035, 0.055, 0.085, 0.135, 0.215, 0.345),
        )
        self.assertEqual(
            tuple(round(x, 3) for x in SELL_LEVELS),
            (0.010, 0.020, 0.035, 0.055, 0.085, 0.135, 0.215),
        )


class GridRoundTripTest(unittest.TestCase):
    """一轮完整买卖：跌 2% 买两档，涨回卖出一档，基准更新。"""

    def test_round_trip_profit(self):
        grid = GridEngine(base_price=100.0, lot_cash=1.0)
        # 第1天：跌到 98（击穿 -1% 和 -2% 两档）
        grid.on_day(0, open_=100.0, high=100.0, low=98.0)
        self.assertEqual(len(grid.lots), 2)
        self.assertAlmostEqual(grid.lots[0][0], 99.0)   # -1% 档
        self.assertAlmostEqual(grid.lots[1][0], 98.0)   # -2% 档
        self.assertFalse(grid.closed)
        # 第2天：涨回 101（击穿 +1% 卖档，卖出 FIFO 最早档）
        grid.on_day(0, open_=99.0, high=101.5, low=99.0)
        self.assertEqual(len(grid.lots), 1)
        # 卖出 99.0 买入的档 @101.0，基准更新为 101.0
        self.assertAlmostEqual(grid.base, 101.0)
        self.assertAlmostEqual(grid.buy_ptr, 0)          # 档位重置
        # 已实现盈亏：份额 = 1/99，卖价 101 -> (101-99)/99
        self.assertAlmostEqual(grid.realized_pnl, 2.0 / 99.0, places=6)

    def test_sell_resets_pointers(self):
        """卖出后基准更新，买入档按新基准重算。"""
        grid = GridEngine(base_price=100.0, lot_cash=1.0)
        grid.on_day(0, 100.0, 100.0, 98.0)      # 买两档 99/98
        grid.on_day(0, 99.0, 101.5, 99.0)       # 卖一档 @101，基准=101
        # 新买档（基准 101）：99.99 / 98.98 / 97.465
        grid.on_day(0, 100.5, 100.5, 98.5)      # 击穿前两档
        self.assertEqual(len(grid.lots), 3)  # 原 98 档 + 两新档
        self.assertAlmostEqual(grid.lots[-1][0], 98.98, places=4)


class SelfProtectionTest(unittest.TestCase):
    def test_three_consecutive_buys_close(self):
        """单日连破三档（累计 -3.5%）触发保护关闭。"""
        grid = GridEngine(base_price=100.0, lot_cash=1.0)
        grid.on_day(0, open_=100.0, high=100.0, low=96.0)
        self.assertEqual(len(grid.lots), 3)
        self.assertTrue(grid.closed, "连续 3 次买入应触发自我保护")

    def test_interleaved_trades_reset_counter(self):
        """买卖穿插（正常震荡）不触发保护。"""
        grid = GridEngine(base_price=100.0, lot_cash=1.0)
        grid.on_day(0, 100.0, 100.0, 98.0)       # 买 2 档
        grid.on_day(0, 99.0, 101.5, 99.0)        # 卖 1 档（计数重置，基准=101）
        # 新档（基准 101）：99.99 / 98.98——low 98.5 只击穿两档，不触发保护
        grid.on_day(0, 100.5, 100.5, 98.5)
        self.assertFalse(grid.closed)
        self.assertEqual(grid.consec_buy, 2)

    def test_force_close_realizes_loss(self):
        grid = GridEngine(base_price=100.0, lot_cash=1.0)
        grid.on_day(0, 100.0, 100.0, 98.0)       # 买 99/98 两档
        grid.force_close(97.0)                # 深跌清仓
        self.assertEqual(len(grid.lots), 0)
        self.assertTrue(grid.closed)
        # 亏损：(97-99)/99 + (97-98)/98
        expected = (97.0 - 99.0) / 99.0 + (97.0 - 98.0) / 98.0
        self.assertAlmostEqual(grid.realized_pnl, expected, places=6)


class GapOpenTest(unittest.TestCase):
    def test_gap_open_fills_at_open(self):
        """跳空低开直接低于档位价时以开盘价成交（更优）。"""
        grid = GridEngine(base_price=100.0, lot_cash=1.0)
        # 开盘 98.5 已低于 -2% 档（98.0）：两档都以 98.5 成交？
        # 触价语义：-1% 档（99.0）以 open=98.5 成交；-2% 档（98.0）
        # low 97.5 击穿，仍以 min(open, level)=98.5？不——开盘后价格
        # 从 98.5 跌到 97.5，第二档以档位价 98.0 成交。
        grid.on_day(0, open_=98.5, high=98.6, low=97.5)
        self.assertAlmostEqual(grid.lots[0][0], 98.5)   # 第一档跳空以 open 成交
        self.assertAlmostEqual(grid.lots[1][0], 98.0)   # 第二档按档位价


if __name__ == "__main__":
    unittest.main()
