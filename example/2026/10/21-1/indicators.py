"""技术指标计算：金叉死叉信号 + 市场状态判定。

纯逻辑模块：只认 pandas Series，不碰网络与文件，与数据周期无关——
日线、60 分钟线、1 分钟线均可，只要序列按时间升序排列，喂什么周期
的数据就算什么周期的指标。

金叉/死叉定义（A = 快线，B = 慢线）：

- 金叉（golden）：昨日 A - B <= 0 且今日 A - B > 0（上穿）
- 死叉（death） ：昨日 A - B >= 0 且今日 A - B < 0（下穿）

必须是"穿越"动作：单纯大于或小于、以及持续粘连都不产生信号。

市场状态判定（regime）指标：

- adx          趋向强度：ADX<25 震荡 / >35 强趋势；+DI/-DI 指示方向
- ma_cohesion  多均线粘合度：<3% 视为粘合（变盘前兆/震荡特征）
- donchian     近期高低点区间：价格是否在震荡区间内运行
- volume_ratio 量比：>1.5 放量（突破有效性确认）
- bias         乖离率：|BIAS|>15% 预警
- drawdown     回撤序列：底仓风控（从最高点回撤超 30% 减持）的原料
"""

from __future__ import annotations

import pandas as pd

GOLDEN = "golden"
DEATH = "death"

__all__ = [
    "GOLDEN",
    "DEATH",
    "ma",
    "ema",
    "wilder_smooth",
    "macd",
    "adx",
    "cross_events",
    "ma_signals",
    "macd_signals",
    "ma_cohesion",
    "donchian",
    "volume_ratio",
    "bias",
    "drawdown",
]


def ma(series: pd.Series, window: int) -> pd.Series:
    """简单移动平均（SMA）。前 window-1 个位置为 NaN：数据不足不给均值。"""
    if window < 1:
        raise ValueError(f"window 必须 >= 1，收到 {window}")
    return series.rolling(window=window, min_periods=window).mean()


def ema(series: pd.Series, window: int) -> pd.Series:
    """
    指数移动平均（EMA），递推式与国内行情软件一致：

        EMA[0] = x[0]
        EMA[t] = alpha * x[t] + (1 - alpha) * EMA[t-1],  alpha = 2 / (window + 1)

    对应 pandas 的 ``ewm(span=window, adjust=False)``。
    """
    if window < 1:
        raise ValueError(f"window 必须 >= 1，收到 {window}")
    return series.ewm(span=window, adjust=False).mean()


def wilder_smooth(series: pd.Series, window: int) -> pd.Series:
    """
    Wilder 平滑（通达信 SMA(X,N,1) 口径）：

        Y[t] = (X[t] + (N-1) * Y[t-1]) / N,  Y[0] = X[0]

    即 ``ewm(alpha=1/N, adjust=False)``。ADX/ATR 等 Wilder 系指标的
    标准平滑方式，与国内行情软件一致。
    """
    if window < 1:
        raise ValueError(f"window 必须 >= 1，收到 {window}")
    return series.ewm(alpha=1.0 / window, adjust=False).mean()


def macd(
    close: pd.Series,
    fast: int = 12,
    slow: int = 26,
    signal: int = 9,
) -> tuple[pd.Series, pd.Series, pd.Series]:
    """
    MACD 三件套，返回 (DIF, DEA, 柱)。

    - DIF = EMA(close, fast) - EMA(close, slow)   快慢线离差
    - DEA = EMA(DIF, signal)                      DIF 的信号线
    - 柱  = 2 * (DIF - DEA)                       国内行情软件惯例乘 2
    """
    if fast >= slow:
        raise ValueError(f"fast({fast}) 必须小于 slow({slow})")
    dif = ema(close, fast) - ema(close, slow)
    dea = ema(dif, signal)
    hist = 2.0 * (dif - dea)
    return dif, dea, hist


def cross_events(fast: pd.Series, slow: pd.Series) -> pd.Series:
    """
    快线上穿/下穿慢线的事件检测。

    返回与 ``fast`` 等长等索引的 object Series：

    - 金叉位置值为 ``"golden"``，死叉位置值为 ``"death"``，其余为 ``pd.NA``

    说明：

    - 输入必须按时间升序排列（``shift(1)`` 依赖行的先后关系）；
    - 任一侧为 NaN 的位置（如均线未成型的前几根）比较结果为 False，
      天然不产生信号；
    - 昨日恰好相等、今日转为正/负，也算穿越（prev<=0 / prev>=0 含等号）。
    """
    if len(fast) != len(slow) or not fast.index.equals(slow.index):
        raise ValueError("fast 与 slow 必须等长且索引一致")

    diff = fast - slow
    prev = diff.shift(1)

    golden = (prev <= 0) & (diff > 0)
    death = (prev >= 0) & (diff < 0)

    events = pd.Series(pd.NA, index=fast.index, dtype="object")
    events[golden] = GOLDEN
    events[death] = DEATH
    return events


def ma_signals(
    close: pd.Series,
    short_window: int = 5,
    long_window: int = 20,
) -> pd.Series:
    """
    均线金叉死叉：短周期 MA 上穿/下穿长周期 MA，默认 MA5 x MA20。

    分钟线数据同样适用：MA5 即"最近 5 根分钟 K 线的均值"。
    """
    if short_window >= long_window:
        raise ValueError(
            f"short_window({short_window}) 必须小于 long_window({long_window})"
        )
    return cross_events(ma(close, short_window), ma(close, long_window))


def macd_signals(
    close: pd.Series,
    fast: int = 12,
    slow: int = 26,
    signal: int = 9,
) -> pd.Series:
    """MACD 金叉死叉：DIF 上穿/下穿 DEA。"""
    dif, dea, _ = macd(close, fast, slow, signal)
    return cross_events(dif, dea)


# ---------------------------------------------------------------------------
# 市场状态判定（regime）指标
# ---------------------------------------------------------------------------


def adx(
    high: pd.Series,
    low: pd.Series,
    close: pd.Series,
    window: int = 14,
) -> tuple[pd.Series, pd.Series, pd.Series]:
    """
    DMI/ADX 趋向指标，返回 (+DI, -DI, ADX)。

    - TR 真实波幅：max(高-低, |高-昨收|, |低-昨收|)，首日回落到当日振幅
    - +DM/-DM 方向变动：仅当单边变动占优且为正时取值，否则记 0
      （+DM = 今日高点抬升幅度，-DM = 今日低点下探幅度）
    - +DI = 100 * WSmooth(+DM) / WSmooth(TR)，-DI 同理
    - DX  = 100 * |+DI - (-DI)| / (+DI + (-DI))，方向变动为零时记 NaN
    - ADX = WSmooth(DX)

    判读（经验阈值）：ADX > 35 强趋势加速；(25, 35] 过渡；< 25 震荡。
    +DI > -DI 看多方向占优，反之看空。
    """
    if window < 1:
        raise ValueError(f"window 必须 >= 1，收到 {window}")
    for name, s in (("high", high), ("low", low)):
        if len(s) != len(close) or not s.index.equals(close.index):
            raise ValueError(f"{name} 与 close 必须等长且索引一致")

    prev_high, prev_low, prev_close = high.shift(1), low.shift(1), close.shift(1)

    up_move = high - prev_high  # 今日向上变动
    down_move = prev_low - low  # 今日向下变动

    plus_dm = up_move.where((up_move > down_move) & (up_move > 0), 0.0)
    minus_dm = down_move.where((down_move > up_move) & (down_move > 0), 0.0)

    tr = pd.concat(
        [high - low, (high - prev_close).abs(), (low - prev_close).abs()],
        axis=1,
    ).max(axis=1)

    smooth_tr = wilder_smooth(tr, window)
    pdi = 100.0 * wilder_smooth(plus_dm, window) / smooth_tr
    mdi = 100.0 * wilder_smooth(minus_dm, window) / smooth_tr

    denom = pdi + mdi
    dx = (100.0 * (pdi - mdi).abs() / denom).where(denom > 0)
    return pdi, mdi, wilder_smooth(dx, window)


def ma_cohesion(
    close: pd.Series,
    windows: tuple[int, ...] = (5, 10, 20),
) -> pd.Series:
    """
    多均线粘合度：(max(MA) - min(MA)) / mean(MA)。

    经验阈值：< 0.03（3%）视为粘合（变盘前兆 / 震荡特征），
    > 0.05 视为发散（趋势展开）。所有均线成型（无 NaN）后才开始计算。
    """
    if len(windows) < 2:
        raise ValueError("windows 至少需要两个才能度量粘合")
    mas = pd.concat([ma(close, w) for w in windows], axis=1)
    ready = mas.notna().all(axis=1)
    spread = (mas.max(axis=1) - mas.min(axis=1)) / mas.mean(axis=1)
    return spread.where(ready)


def donchian(
    high: pd.Series,
    low: pd.Series,
    window: int = 20,
) -> tuple[pd.Series, pd.Series]:
    """
    Donchian 通道，返回 (上轨, 下轨)：

    - 上轨 = 前 window 根K线的最高价（不含当日）
    - 下轨 = 前 window 根K线的最低价（不含当日）

    shift(1) 排除当日是刻意的：若把当日算入，当日创新高时上轨
    恒等于当日高点，"突破上轨"永远无法成立。用途：震荡区间判定、
    前期高点压力位 / 低点支撑位。
    """
    if window < 1:
        raise ValueError(f"window 必须 >= 1，收到 {window}")
    upper = high.rolling(window).max().shift(1)
    lower = low.rolling(window).min().shift(1)
    return upper, lower


def volume_ratio(volume: pd.Series, window: int = 5) -> pd.Series:
    """
    量比：当日成交量 / 前 window 日平均成交量（不含当日）。

    经验阈值：> 1.5 放量（突破有效性确认），< 0.7 缩量。
    """
    if window < 1:
        raise ValueError(f"window 必须 >= 1，收到 {window}")
    return volume / volume.rolling(window).mean().shift(1)


def bias(close: pd.Series, window: int = 10) -> pd.Series:
    """
    乖离率 BIAS = (close - MA) / MA。

    衡量价格偏离均线的程度，经验阈值：|BIAS| > 15% 预警。
    """
    if window < 1:
        raise ValueError(f"window 必须 >= 1，收到 {window}")
    m = ma(close, window)
    return (close - m) / m


def drawdown(close: pd.Series) -> pd.Series:
    """
    回撤序列：close / 历史最高 - 1，恒 <= 0。

    底仓风控原料：从最高点回撤超过 30% 触发减持。
    """
    return close / close.cummax() - 1.0
