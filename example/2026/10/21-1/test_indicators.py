"""indicators.py 的单元测试：构造数据验证交叉检测，不依赖网络。

运行方式（在本目录下）：

    python3 -m unittest test_indicators -v
"""

import math
import unittest

import pandas as pd

from indicators import (
    DEATH,
    GOLDEN,
    adx,
    bias,
    cross_events,
    donchian,
    drawdown,
    ema,
    ma,
    ma_cohesion,
    macd,
    macd_divergence,
    macd_signals,
    ma_signals,
    volume_ratio,
    wilder_smooth,
)


class MaTest(unittest.TestCase):
    def test_basic(self):
        close = pd.Series([1.0, 2.0, 3.0, 4.0, 5.0])
        result = ma(close, 3)
        # 前 window-1 个位置为 NaN：数据不足不给均值
        self.assertTrue(result.iloc[:2].isna().all())
        self.assertAlmostEqual(result.iloc[2], 2.0)
        self.assertAlmostEqual(result.iloc[3], 3.0)
        self.assertAlmostEqual(result.iloc[4], 4.0)

    def test_invalid_window(self):
        with self.assertRaises(ValueError):
            ma(pd.Series([1.0, 2.0]), 0)


class EmaTest(unittest.TestCase):
    def test_recursive_formula(self):
        # window=2, alpha=2/3，手工递推验证：
        #   EMA[0] = 10
        #   EMA[1] = 2/3*20 + 1/3*10 = 50/3
        #   EMA[2] = 2/3*30 + 1/3*(50/3) = 230/9
        series = pd.Series([10.0, 20.0, 30.0])
        result = ema(series, 2)
        self.assertAlmostEqual(result.iloc[0], 10.0)
        self.assertAlmostEqual(result.iloc[1], 50.0 / 3.0)
        self.assertAlmostEqual(result.iloc[2], 230.0 / 9.0)

    def test_invalid_window(self):
        with self.assertRaises(ValueError):
            ema(pd.Series([1.0, 2.0]), -1)


class MacdTest(unittest.TestCase):
    def test_definition(self):
        # 固定形态的数据：先跌后涨，保证 EMA 有明显变化
        close = pd.Series(
            [100.0 - i for i in range(40)] + [60.0 + i for i in range(40)]
        )
        dif, dea, hist = macd(close)
        expected_dif = ema(close, 12) - ema(close, 26)
        expected_dea = ema(expected_dif, 9)
        pd.testing.assert_series_equal(dif, expected_dif)
        pd.testing.assert_series_equal(dea, expected_dea)
        pd.testing.assert_series_equal(hist, 2.0 * (dif - dea))

    def test_invalid_params(self):
        with self.assertRaises(ValueError):
            macd(pd.Series([1.0, 2.0, 3.0]), fast=26, slow=12)


class CrossEventsTest(unittest.TestCase):
    def test_golden_and_death(self):
        # diff = [-1, -1, +1, +1, -1, -1]：
        # 位置 2 昨日 -1 今日 +1 -> 金叉；位置 4 昨日 +1 今日 -1 -> 死叉
        fast = pd.Series([1.0, 1.0, 3.0, 3.0, 1.0, 1.0])
        slow = pd.Series([2.0] * 6)
        events = cross_events(fast, slow)
        # 注：pandas 的 pd.NA 与 None 互不相等，须逐位置断言
        self.assertEqual(events.iloc[2], GOLDEN)
        self.assertEqual(events.iloc[4], DEATH)
        self.assertTrue(events.iloc[[0, 1, 3, 5]].isna().all())

    def test_cross_through_equal_counts(self):
        # 昨日恰好相等、今日转正/转负也算穿越：
        # diff = [-1, 0, +1, 0, 0, -1]
        fast = pd.Series([1.0, 2.0, 3.0, 2.0, 2.0, 1.0])
        slow = pd.Series([2.0] * 6)
        events = cross_events(fast, slow)
        self.assertEqual(events.iloc[2], GOLDEN)
        self.assertEqual(events.iloc[5], DEATH)
        self.assertTrue(events.iloc[[0, 1, 3, 4]].isna().all())

    def test_consistently_above_no_signal(self):
        # 持续大于不算交叉
        fast = pd.Series([3.0, 4.0, 5.0, 6.0])
        slow = pd.Series([2.0] * 4)
        events = cross_events(fast, slow)
        self.assertTrue(events.isna().all())

    def test_nan_region_no_signal(self):
        # 首行 prev 为 NaN，即使今日 diff > 0 也不产生信号
        fast = pd.Series([3.0, 4.0, 5.0])
        slow = pd.Series([2.0] * 3)
        events = cross_events(fast, slow)
        self.assertTrue(events.isna().all())

    def test_index_mismatch_raises(self):
        fast = pd.Series([1.0, 2.0, 3.0])
        slow = pd.Series([1.0, 2.0, 3.0], index=[0, 1, 3])
        with self.assertRaises(ValueError):
            cross_events(fast, slow)


class MaSignalsTest(unittest.TestCase):
    def test_v_shape_price(self):
        # 横盘让 MA20 成型 -> 上涨触发金叉 -> 回落触发死叉
        close = pd.Series(
            [100.0] * 25
            + [100.0 + 2.0 * i for i in range(1, 16)]   # 100 -> 130
            + [130.0 - 2.0 * i for i in range(1, 16)]    # 130 -> 100
        )
        signals = ma_signals(close, 5, 20)

        self.assertEqual(signals.dropna().iloc[0], GOLDEN, "上涨段应先出现金叉")
        self.assertEqual(signals.dropna().iloc[-1], DEATH, "回落段应最终出现死叉")

    def test_ma_warmup_all_na(self):
        # 前 long_window-1 根长均线为 NaN，该区域不应产生任何信号
        close = pd.Series([100.0 + i for i in range(50)])
        signals = ma_signals(close, 5, 20)
        self.assertTrue(signals.iloc[:19].isna().all())

    def test_invalid_windows(self):
        close = pd.Series([1.0] * 30)
        with self.assertRaises(ValueError):
            ma_signals(close, short_window=20, long_window=5)


class MacdSignalsTest(unittest.TestCase):
    def test_sine_wave_alternating(self):
        # 周期 20 的正弦波：DIF/DEA 应反复交叉，且金叉死叉严格交替
        close = pd.Series(
            [100.0 + 5.0 * math.sin(i * math.pi / 10.0) for i in range(120)]
        )
        signals = macd_signals(close).dropna()

        self.assertGreaterEqual(len(signals), 2, "正弦波应产生多组交叉")
        kinds = signals.tolist()
        for prev_kind, curr_kind in zip(kinds, kinds[1:]):
            self.assertNotEqual(
                prev_kind, curr_kind, "相邻信号必须金叉死叉交替"
            )
        # 从正弦波起点开始先涨，首个信号应为金叉
        self.assertEqual(kinds[0], GOLDEN)


class WilderSmoothTest(unittest.TestCase):
    def test_recursive_formula(self):
        # window=3, alpha=1/3，手工递推：
        #   Y[0]=9, Y[1]=(6+2*9)/3=8, Y[2]=(12+2*8)/3=28/3
        series = pd.Series([9.0, 6.0, 12.0])
        result = wilder_smooth(series, 3)
        self.assertAlmostEqual(result.iloc[0], 9.0)
        self.assertAlmostEqual(result.iloc[1], 8.0)
        self.assertAlmostEqual(result.iloc[2], 28.0 / 3.0)

    def test_invalid_window(self):
        with self.assertRaises(ValueError):
            wilder_smooth(pd.Series([1.0]), 0)


class AdxTest(unittest.TestCase):
    def test_uptrend_bullish(self):
        # 单调上涨：每日高点/低点同步抬升 -> +DM 持续为正, -DM 归零
        n = 60
        close = pd.Series([100.0 + i for i in range(n)])
        high, low = close + 1.0, close - 1.0
        pdi, mdi, adx_line = adx(high, low, close, 14)
        self.assertGreater(pdi.iloc[-1], mdi.iloc[-1], "上涨趋势 +DI 应压倒 -DI")
        self.assertGreater(adx_line.iloc[-1], 25.0, "单边趋势 ADX 应 > 25")

    def test_downtrend_bearish(self):
        n = 60
        close = pd.Series([160.0 - i for i in range(n)])
        high, low = close + 1.0, close - 1.0
        pdi, mdi, adx_line = adx(high, low, close, 14)
        self.assertGreater(mdi.iloc[-1], pdi.iloc[-1], "下跌趋势 -DI 应压倒 +DI")
        self.assertGreater(adx_line.iloc[-1], 25.0)

    def test_sideways_low_directional_gap(self):
        # 横盘：涨跌交替 -> +DI/-DI 接近，ADX 低位
        n = 60
        close = pd.Series([100.0 if i % 2 == 0 else 100.1 for i in range(n)])
        high, low = close + 0.05, close - 0.05
        pdi, mdi, adx_line = adx(high, low, close, 14)
        self.assertLess(abs(pdi.iloc[-1] - mdi.iloc[-1]), 5.0)
        self.assertLess(adx_line.iloc[-1], 25.0)

    def test_index_mismatch_raises(self):
        close = pd.Series([10.0, 11.0, 12.0])
        high = pd.Series([11.0, 12.0, 13.0], index=[0, 1, 3])
        with self.assertRaises(ValueError):
            adx(high, close - 1.0, close)


class MaCohesionTest(unittest.TestCase):
    def test_flat_is_tight(self):
        close = pd.Series([100.0] * 30)
        self.assertLess(ma_cohesion(close).iloc[-1], 0.03)

    def test_trend_is_loose(self):
        close = pd.Series([100.0 + 2.0 * i for i in range(30)])
        self.assertGreater(ma_cohesion(close).iloc[-1], 0.05)

    def test_warmup_is_nan(self):
        # 最长窗口 20 未成型前（前 19 根）不应给出粘合度
        close = pd.Series([100.0 + i for i in range(30)])
        self.assertTrue(ma_cohesion(close).iloc[:19].isna().all())

    def test_single_window_raises(self):
        with self.assertRaises(ValueError):
            ma_cohesion(pd.Series([1.0] * 30), windows=(5,))


class DonchianTest(unittest.TestCase):
    def test_basic_values(self):
        high = pd.Series([1.0, 2.0, 3.0, 4.0, 5.0, 6.0])
        low = pd.Series([0.5, 1.0, 1.5, 2.0, 2.5, 3.0])
        upper, lower = donchian(high, low, 3)
        # 第4根(t=3)的上轨 = 前3根最高 = 3，下轨 = 前3根最低 = 0.5
        self.assertAlmostEqual(upper.iloc[3], 3.0)
        self.assertAlmostEqual(lower.iloc[3], 0.5)
        self.assertAlmostEqual(upper.iloc[5], 5.0)
        self.assertTrue(upper.iloc[:3].isna().all())

    def test_excludes_today(self):
        # 当日创新高不应抬高当日上轨——否则"突破上轨"永远无法成立
        high = pd.Series([1.0, 2.0, 3.0, 4.0, 5.0, 100.0])
        low = pd.Series([0.5] * 6)
        upper, _ = donchian(high, low, 3)
        self.assertAlmostEqual(upper.iloc[5], 5.0)


class VolumeRatioTest(unittest.TestCase):
    def test_basic(self):
        vol = pd.Series([100.0, 200.0, 300.0, 400.0, 500.0, 1000.0])
        vr = volume_ratio(vol, 5)
        # 尾值 = 1000 / mean(100..500) = 1000/300
        self.assertAlmostEqual(vr.iloc[-1], 1000.0 / 300.0)
        self.assertTrue(vr.iloc[:5].isna().all())

    def test_invalid_window(self):
        with self.assertRaises(ValueError):
            volume_ratio(pd.Series([1.0]), 0)


class BiasTest(unittest.TestCase):
    def test_hand_computed(self):
        close = pd.Series([10.0, 10.0, 10.0, 13.0])
        b = bias(close, 3)
        # MA3[3] = (10+10+13)/3 = 11, BIAS = (13-11)/11 = 2/11
        self.assertAlmostEqual(b.iloc[3], 2.0 / 11.0)
        self.assertTrue(b.iloc[:2].isna().all())

    def test_sign_follows_trend(self):
        up = pd.Series([100.0 + i for i in range(30)])
        down = pd.Series([130.0 - i for i in range(30)])
        self.assertGreater(bias(up).iloc[-1], 0.0)
        self.assertLess(bias(down).iloc[-1], 0.0)


class DrawdownTest(unittest.TestCase):
    def test_hand_computed(self):
        close = pd.Series([100.0, 120.0, 90.0])
        dd = drawdown(close)
        self.assertAlmostEqual(dd.iloc[0], 0.0)
        self.assertAlmostEqual(dd.iloc[1], 0.0)
        self.assertAlmostEqual(dd.iloc[2], 90.0 / 120.0 - 1.0)

    def test_never_positive(self):
        close = pd.Series([100.0, 110.0, 95.0, 130.0, 80.0, 125.0])
        self.assertTrue((drawdown(close) <= 0.0).all())

    def test_new_high_resets(self):
        close = pd.Series([100.0, 120.0, 90.0, 130.0])
        self.assertAlmostEqual(drawdown(close).iloc[-1], 0.0)


class MacdDivergenceTest(unittest.TestCase):
    def _rally_then_spike(self):
        """急涨创高 -> 深回调 -> 短促冲高（末根唯一新高，DIF 未恢复）。"""
        return pd.Series(
            [100.0] * 20                                    # 预热横盘
            + [100.0 + 2.0 * i for i in range(1, 11)]       # 急涨 102..120
            + [120.0 - 2.0 * i for i in range(1, 13)]       # 深回调 118..96
            + [100.0, 106.0, 113.0, 121.0]                  # 短促冲高，末根新高
        )

    def test_bear_divergence_on_new_high_with_weak_dif(self):
        close = self._rally_then_spike()
        dif, _, _ = macd(close)
        # 前置条件：冲高末根价格新高但 DIF 低于第一波峰值
        self.assertGreater(close.iloc[-1], 120.0)
        self.assertLess(dif.iloc[-1], dif.iloc[29], "构造前提：DIF 未恢复到第一波峰值")
        bear, _ = macd_divergence(close, dif, window=20)
        self.assertTrue(bear.iloc[-1], "价格新高+动能未新高应检出顶背离")
        self.assertFalse(bear.iloc[:30].any(), "第一波单调上涨中不应有背离")

    def test_bull_divergence_on_new_low_with_strong_dif(self):
        close = 200.0 - self._rally_then_spike()  # 镜像：急跌/深反弹/短促下破
        dif, _, _ = macd(close)
        _, bull = macd_divergence(close, dif, window=20)
        self.assertTrue(bull.iloc[-1], "价格新低+动能未新低应检出底背离")

    def test_no_divergence_in_accelerating_trend(self):
        # 加速上涨：价格与 DIF 同步新高，不应有顶背离
        close = pd.Series([100.0 * (1.05 ** i) for i in range(40)])
        dif, _, _ = macd(close)
        bear, _ = macd_divergence(close, dif, window=5)
        self.assertFalse(bear.any())

    def test_touching_prev_high_is_not_breakout(self):
        # 触及（==）前高不算突破：横盘触及不产生高点日
        close = pd.Series([100.0] * 10 + [90.0, 100.0] * 5)
        dif, _, _ = macd(close)
        bear, bull = macd_divergence(close, dif, window=3)
        self.assertFalse(bear.any() or bull.any())

    def test_mismatch_raises(self):
        close = pd.Series([1.0, 2.0, 3.0])
        dif = pd.Series([0.1, 0.2, 0.3], index=[0, 1, 3])
        with self.assertRaises(ValueError):
            macd_divergence(close, dif, 5)


if __name__ == "__main__":
    unittest.main()
