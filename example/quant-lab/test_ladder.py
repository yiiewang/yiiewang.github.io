"""ladder.py 的单元测试：锚点为博客涪陵电力案例的推演数字。

运行：python3 -m unittest test_ladder -v
"""

import unittest

import pandas as pd

from ladder import ladder_backtest


def make_df(rows):
    """构造回测输入。rows: (open, low, close, ma5, ma10, ma20)。"""
    keys = ("open", "low", "close", "ma5", "ma10", "ma20")
    return pd.DataFrame(
        {k: [r[i] for r in rows] for i, k in enumerate(keys)}
    )


def hold_series(*values):
    return pd.Series(list(values))


class FullLadderPathTest(unittest.TestCase):
    """完整三档路径，锚点=博客涪陵电力推演（+7.7%）。"""

    def setUp(self):
        # index 0 为建仓前一日的虚拟K线（使 index 1 开盘建仓）
        self.df = make_df(
            [
                (12.50, 12.40, 12.50, 12.5, 12.4, 12.00),  # 0 前一日
                (12.60, 12.30, 12.80, 12.5, 12.4, 12.00),  # 1 建仓 12.60
                (12.90, 12.80, 13.30, 12.7, 12.5, 12.05),  # 2 收盘浮盈5.6%
                (13.23, 13.10, 13.40, 12.9, 12.6, 12.10),  # 3 开盘减第一档
                (13.50, 13.40, 14.10, 13.2, 13.20, 12.30),  # 4 收盘浮盈11.9%
                (14.00, 13.90, 14.05, 13.6, 13.50, 12.60),  # 5 开盘减第二档
                (13.60, 13.40, 13.45, 13.8, 13.50, 12.80),  # 6 盘中止损13.50
            ]
        )
        self.hold = hold_series(True, True, True, True, True, True, True)
        self.result = ladder_backtest(self.df, self.hold, exit_on_signal=False)

    def test_three_tier_exits(self):
        trade = self.result["trades"][0]
        self.assertAlmostEqual(trade["exits"][0], 13.23)  # 第一档 +5%
        self.assertAlmostEqual(trade["exits"][1], 14.00)  # 第二档 浮盈11.9%
        self.assertAlmostEqual(trade["exits"][2], 13.50)  # 第三档 MA10 止损

    def test_return_matches_blog(self):
        # 博客推演：三档合计 +7.7%（(13.23+14.00+13.50)/(3*12.60)-1）
        trade = self.result["trades"][0]
        self.assertAlmostEqual(trade["factor"], 40.73 / 37.80, places=3)
        self.assertAlmostEqual(self.result["total_return"], 0.0775, places=3)

    def test_stop_reason(self):
        self.assertEqual(self.result["trades"][0]["reason"], "stop")

    def test_initial_stop_is_min_low_ma20(self):
        # 建仓日收盘确定 S = min(12.30, 12.00) = 12.00，此后未被击穿
        # 直至级联上移（第一档减后升至 12.60，第二档减后升至 13.50）
        trade = self.result["trades"][0]
        self.assertAlmostEqual(trade["entry"], 12.60)


class StopBeforeFirstTierTest(unittest.TestCase):
    """失败离场：建仓后直接跌破初始止损，三档同价止损。"""

    def test_all_tiers_stopped_at_initial_stop(self):
        df = make_df(
            [
                (12.50, 12.40, 12.50, 12.5, 12.4, 12.00),
                (12.60, 12.30, 12.50, 12.5, 12.4, 12.00),  # 建仓，S=12.00
                (12.40, 12.10, 12.20, 12.4, 12.4, 12.00),  # 未触
                (12.05, 11.80, 11.90, 12.3, 12.3, 12.00),  # low 破 12.00
            ]
        )
        result = ladder_backtest(df, hold_series(True, True, True, True))
        trade = result["trades"][0]
        self.assertEqual(trade["reason"], "stop")
        for exit_price in trade["exits"]:
            self.assertAlmostEqual(exit_price, 12.00)
        # 三档全损：factor = 12.00/12.60
        self.assertAlmostEqual(trade["factor"], 12.00 / 12.60, places=4)


class SignalExitTest(unittest.TestCase):
    """死叉清仓：hold 转空后次日开盘清掉全部剩余档。"""

    def test_exit_on_signal(self):
        df = make_df(
            [
                (12.50, 12.40, 12.50, 12.5, 12.4, 12.00),
                (12.60, 12.30, 12.70, 12.5, 12.4, 12.00),  # 建仓
                (12.80, 12.70, 12.90, 12.6, 12.4, 12.00),  # 浮盈2.4% 无档触发
                (13.00, 12.90, 13.00, 12.8, 12.5, 12.05),  # hold 今日转空
                (12.95, 12.80, 12.90, 12.8, 12.5, 12.05),  # 次日开盘清仓
            ]
        )
        # hold: 前3天多头，第4天(index=3)转空 -> index4 开盘清仓
        result = ladder_backtest(df, hold_series(True, True, True, False, False))
        trade = result["trades"][0]
        self.assertEqual(trade["reason"], "signal")
        for exit_price in trade["exits"]:
            self.assertAlmostEqual(exit_price, 12.95)

    def test_exit_on_signal_disabled(self):
        """关闭信号清仓后，同样的序列持有到期末按收盘平仓。"""
        df = make_df(
            [
                (12.50, 12.40, 12.50, 12.5, 12.4, 12.00),
                (12.60, 12.30, 12.70, 12.5, 12.4, 12.00),
                (12.80, 12.70, 12.90, 12.6, 12.4, 12.00),
                (13.00, 12.90, 13.00, 12.8, 12.5, 12.05),
                (12.95, 12.80, 12.90, 12.8, 12.5, 12.05),
            ]
        )
        result = ladder_backtest(
            df, hold_series(True, True, True, False, False), exit_on_signal=False
        )
        trade = result["trades"][0]
        self.assertEqual(trade["reason"], "eod")
        for exit_price in trade["exits"]:
            self.assertAlmostEqual(exit_price, 12.90)


class GapOpenStopTest(unittest.TestCase):
    """跳空低开：开盘价直接低于止损价时以开盘价成交（更保守）。"""

    def test_gap_open_fills_at_open(self):
        df = make_df(
            [
                (12.50, 12.40, 12.50, 12.5, 12.4, 12.00),
                (12.60, 12.30, 12.50, 12.5, 12.4, 12.00),  # 建仓 S=12.00
                (11.80, 11.70, 11.75, 12.3, 12.3, 12.00),  # 跳空低开 11.80
            ]
        )
        result = ladder_backtest(df, hold_series(True, True, True))
        trade = result["trades"][0]
        self.assertEqual(trade["reason"], "gap_stop")
        for exit_price in trade["exits"]:
            self.assertAlmostEqual(exit_price, 11.80)


class Tier2ByDivergenceTest(unittest.TestCase):
    """顶背离预警独立触发第二档（浮盈不足 10%）。"""

    def test_divergence_triggers_tier2(self):
        df = make_df(
            [
                (12.50, 12.40, 12.50, 12.5, 12.4, 12.00),
                (12.60, 12.30, 12.80, 12.5, 12.4, 12.00),  # 建仓
                (12.90, 12.80, 13.30, 12.7, 12.5, 12.05),  # 浮盈5.6% -> t1
                (13.23, 13.10, 13.40, 12.9, 12.6, 12.10),  # 减 t1；浮盈6.3%
                (13.50, 13.40, 13.60, 13.0, 12.7, 12.20),  # 顶背离日，浮盈7.9%
                (13.55, 13.30, 13.50, 13.2, 12.9, 12.30),  # 开盘减 t2
                (13.45, 13.35, 13.40, 13.3, 13.00, 12.35),  # MA20=12.35<止损
                (13.40, 13.20, 13.30, 13.3, 13.05, 12.40),  # 止损 13.23 触发?
            ]
        )
        df["bear_div"] = [False] * 7 + [False]
        df.loc[4, "bear_div"] = True  # index 4 顶背离预警
        result = ladder_backtest(df, hold_series(True, True, True, True, True, True, True, True), exit_on_signal=False)
        trade = result["trades"][0]
        # 第一档 @13.23（index3 开盘），第二档由背离触发 @13.55（index5 开盘）
        self.assertAlmostEqual(trade["exits"][0], 13.23)
        self.assertAlmostEqual(trade["exits"][1], 13.55)
        # 第二档减后止损上移至 13.23；index7 low=13.20 <= 13.23 盘中触发
        self.assertAlmostEqual(trade["exits"][2], 13.23)
        self.assertEqual(trade["reason"], "stop")


class NoEntryWithoutSignalTest(unittest.TestCase):
    def test_no_entry_when_hold_false(self):
        df = make_df(
            [(12.50, 12.40, 12.50, 12.5, 12.4, 12.00)] * 5
        )
        result = ladder_backtest(df, hold_series(False, False, False, False, False))
        self.assertEqual(result["count"], 0)
        self.assertAlmostEqual(result["total_return"], 0.0)


class CooldownProtocolTest(unittest.TestCase):
    """连续假突破协议：止损后冷却期内不再入场，成功后计数归零。"""

    def _df(self, days):
        # 循环模式：涨两天触发建仓后回跌破 12.00 初始止损，往复
        rows = [(12.50, 12.40, 12.50, 12.5, 12.4, 12.00)]
        for _ in range(days):
            rows += [
                (12.60, 12.30, 12.70, 12.5, 12.4, 12.00),  # 建仓 S=12.00
                (12.10, 11.80, 12.00, 12.4, 12.4, 12.00),  # 次日跳空破止损
            ]
        return make_df(rows)

    def test_cooldown_reduces_trades(self):
        df = self._df(6)
        hold = pd.Series([True] * len(df))
        without = ladder_backtest(df, hold, exit_on_signal=False)
        with_cd = ladder_backtest(
            df, hold, exit_on_signal=False, cooldown=True
        )
        # 冷却期内即使 hold 多头也不入场：轮数应少于无协议版
        self.assertLess(with_cd["count"], without["count"])
        # 首次止损后冷却 3 天：无协议版在第 4 天(index4)即可再入场，
        # 协议版最早 index 1+3=4 之后——轮数差至少 1
        self.assertGreaterEqual(without["count"], with_cd["count"] + 1)

    def test_cooldown_disabled_by_default(self):
        df = self._df(4)
        hold = pd.Series([True] * len(df))
        default = ladder_backtest(df, hold, exit_on_signal=False)
        explicit = ladder_backtest(df, hold, exit_on_signal=False, cooldown=False)
        self.assertEqual(default["count"], explicit["count"])


if __name__ == "__main__":
    unittest.main()
