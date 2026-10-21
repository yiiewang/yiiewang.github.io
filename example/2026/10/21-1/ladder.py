"""三档阶梯止盈回测状态机（复合系统 V2.1 的可回测实现）。

规则翻译自《当趋势策略遇上震荡市》三档阶梯止盈策略：

- **建仓拆档**：hold 昨日为 True 且当前空仓 -> 今日开盘建仓，
  仓位按 33%/33%/34% 拆三档（实现取等权 1/3）
- **初始止损**：入场日收盘确定 S = min(入场日 low, 入场日 MA20)，
  次日起挂单——入场日当天裸露（忠实原规则且无未来函数：当日
  low 收盘才可知）
- **第一档（安心仓）**：收盘浮盈 >= 5% -> 次日开盘减仓，
  减后止损上移至成本价
- **第二档（利润仓）**：收盘浮盈 >= 10% 或 顶背离预警 -> 次日开盘
  减仓，减后止损上移至第一档减仓价
- **第三档（趋势仓）**：纯均线移动止损（只上不下）：第一档已减后
  = MA20，浮盈 10-20% = MA10，浮盈 > 20% = MA5

执行语义（无未来函数）：

- 建仓/减仓/信号清仓：T 收盘判定、T+1 开盘成交
- 止损：挂单盘中触价成交（当日 low <= stop；跳空低开时以 open 成交）
- exit_on_signal=True：hold 转空 -> 次日开盘清仓全部剩余档（保守版，
  可关闭以测试纯移动止损）
"""

from __future__ import annotations

import pandas as pd

TIER1_TAKE = 0.05  # 第一档触发浮盈
TIER2_TAKE = 0.10  # 第二档触发浮盈


def ladder_backtest(
    df: pd.DataFrame,
    hold: pd.Series,
    exit_on_signal: bool = True,
) -> dict:
    """
    三档阶梯止盈回测。df 需含 open/close/low/ma5/ma10/ma20 列，
    可选 bear_div 列（第二档预警条件）。

    返回：
    - trades: 每轮明细 [{entry, exits(三档出场价), reason, factor}]
    - count / win_rate / total_return / max_dd / equity_curve
    """
    required = {"open", "close", "low", "ma5", "ma10", "ma20"}
    if not required.issubset(df.columns):
        raise ValueError(f"df 缺少必需列: {sorted(required - set(df.columns))}")

    opens = df["open"].astype(float).tolist()
    lows = df["low"].astype(float).tolist()
    closes = df["close"].astype(float).tolist()
    ma5 = df["ma5"].astype(float).tolist()
    ma10 = df["ma10"].astype(float).tolist()
    ma20 = df["ma20"].astype(float).tolist()
    bear_div = (
        df["bear_div"].fillna(False).astype(bool).tolist()
        if "bear_div" in df.columns
        else [False] * len(df)
    )
    n = len(df)
    should_enter = hold.shift(1).fillna(False).astype(bool).tolist()
    should_exit = (
        (~hold).shift(1).fillna(False).astype(bool).tolist()
        if exit_on_signal
        else [False] * n
    )

    trades: list[dict] = []
    equity = 1.0   # 全局净值（轮间复利）
    base = 1.0     # 本轮起点的全局净值
    entry = 0.0
    stop = 0.0
    tier_open = [False, False, False]
    tier_exit: list[float | None] = [None, None, None]
    holding = False
    entry_day = -1
    pending: list[str] = []
    equity_curve: list[float] = [1.0]

    def round_value(price: float) -> float:
        """本轮相对价值：已减档按出场价锁定，在场档按 price 估值。"""
        total = 0.0
        for t in range(3):
            if tier_exit[t] is not None:
                total += tier_exit[t] / entry
            else:
                total += price / entry
        return total / 3.0

    def finalize(reason: str) -> None:
        nonlocal holding, base, equity, pending
        factor = sum(x / entry for x in tier_exit) / 3.0
        equity = base * factor
        trades.append(
            {
                "entry": entry,
                "exits": tuple(tier_exit),
                "reason": reason,
                "factor": factor,
            }
        )
        holding = False
        pending = []
        base = equity

    def close_remaining(price: float, reason: str) -> None:
        nonlocal tier_open
        for t in range(3):
            if tier_open[t]:
                tier_exit[t] = price
                tier_open[t] = False
        finalize(reason)

    for i in range(n):
        # ---------- 开盘阶段：执行昨日收盘判定的动作 ----------
        if holding:
            if should_exit[i]:
                close_remaining(opens[i], "signal")
            elif opens[i] <= stop > 0:
                close_remaining(opens[i], "gap_stop")
            elif pending:
                if "t1" in pending and tier_open[0]:
                    tier_exit[0] = opens[i]
                    tier_open[0] = False
                if "t2" in pending and tier_open[1]:
                    tier_exit[1] = opens[i]
                    tier_open[1] = False
                pending = []
                if not any(tier_open):
                    finalize("tiers")
        elif should_enter[i]:
            entry = opens[i]
            tier_open = [True, True, True]
            tier_exit = [None, None, None]
            holding = True
            entry_day = i
            stop = 0.0
            pending = []

        # ---------- 盘中阶段：触价止损 ----------
        if holding and stop > 0 and opens[i] > stop and lows[i] <= stop:
            close_remaining(stop, "stop")

        # ---------- 收盘阶段：判定与止损上移（次日起生效） ----------
        if holding:
            if i == entry_day:
                stop = min(lows[i], ma20[i])
            else:
                gain = closes[i] / entry - 1.0
                if tier_open[0] and gain >= TIER1_TAKE:
                    pending.append("t1")
                if tier_open[1] and (gain >= TIER2_TAKE or bear_div[i]):
                    pending.append("t2")
                # 级联上移：只上不下
                if not tier_open[0]:
                    stop = max(stop, entry)          # 第一档减后 -> 成本价
                if not tier_open[1] and tier_exit[0] is not None:
                    stop = max(stop, tier_exit[0])   # 第二档减后 -> 第一档价
                # 第三档移动止损（均线梯度）
                if not tier_open[0]:
                    stop = max(stop, ma20[i])
                if gain >= 0.10:
                    stop = max(stop, ma10[i])
                if gain > 0.20:
                    stop = max(stop, ma5[i])

        # ---------- 净值 ----------
        if holding:
            equity = base * round_value(closes[i])
        equity_curve.append(equity)

    # 期末持仓按收盘平仓
    if holding:
        close_remaining(closes[-1], "eod")
        equity_curve.append(equity)

    eq = pd.Series(equity_curve)
    return {
        "trades": trades,
        "count": len(trades),
        "win_rate": (
            sum(1 for t in trades if t["factor"] > 1.0) / len(trades)
            if trades
            else None
        ),
        "total_return": equity - 1.0,
        "max_dd": float((eq / eq.cummax() - 1.0).min()),
        "equity_curve": eq,
    }
