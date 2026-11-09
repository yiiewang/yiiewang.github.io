"""斐波那契递增网格引擎（复合系统网格层的可回测实现）。

设计翻译自《当趋势策略遇上震荡市》网格层：

- **递增间距**：单档间距 1, 1, 1.5, 2, 3, 5, 8, 13（%）近似斐波那契——
  浅层密集收割震荡，深层稀疏保留子弹（越跌越疏，防单边深套）
- **买入档**（累计跌幅）：-1, -2, -3.5, -5.5, -8.5, -13.5, -21.5, -34.5%
- **卖出档**（累计涨幅）：+1, +2, +3.5, +5.5, +8.5, +13.5, +21.5%
  （卖 7 档，深层仓位移交趋势层——本实现不建模移交，卖完即止）
- **基准价更新**：每次卖出后基准更新为卖出价，档位指针重置
- **自我保护**：连续 3 次同向成交（无反向穿插）判定单边化，
  引擎关闭，剩余持仓按收盘清仓

成交语义：触价成交（low <= 买档价 / high >= 卖档价），跳空时以
开盘价成交。网格死于单边：趋势市里买入档连环触发（越买越套），
本引擎的关闭机制是最后一道保险——真正的防护靠切换层（regime）。
"""

from __future__ import annotations

import pandas as pd

# 单档间距（近似斐波那契数列）
FIB_STEPS = (0.010, 0.010, 0.015, 0.020, 0.030, 0.050, 0.080, 0.130)


def _cumsum(steps: tuple[float, ...]) -> tuple[float, ...]:
    out, total = [], 0.0
    for s in steps:
        total += s
        out.append(total)
    return tuple(out)


BUY_LEVELS = _cumsum(FIB_STEPS)          # 8 档累计跌幅
SELL_LEVELS = _cumsum(FIB_STEPS[:7])     # 7 档累计涨幅

MAX_CONSECUTIVE = 3  # 连续同向成交的保护阈值


class GridEngine:
    """单次运行的网格引擎：档位推进、成交配对、自我保护。"""

    def __init__(self, base_price: float, lot_cash: float):
        self.base = base_price
        self.lot_cash = lot_cash          # 每档投入资金
        self.buy_ptr = 0                  # 下一个未触发买档
        self.sell_ptr = 0                 # 下一个未触发卖档
        self.lots: list[tuple[float, float]] = []  # (买入价, 份额)
        self.consec_buy = 0
        self.consec_sell = 0
        self.closed = False               # 自我保护触发
        self.cash_out = 0.0               # 已实现卖出入金
        self.realized_pnl = 0.0           # 已实现盈亏（相对买入成本）
        self.invested = 0.0               # 累计买入资金
        self.fills: list[dict] = []       # 成交记录 [{i,a,p}]

    def on_day(self, i: int, open_: float, high: float, low: float) -> None:
        """推进一日：按触价档位成交，检查自我保护。"""
        if self.closed:
            return
        # 买入：low 击穿买档
        while self.buy_ptr < len(BUY_LEVELS) and not self.closed:
            level = self.base * (1.0 - BUY_LEVELS[self.buy_ptr])
            if low <= level:
                price = min(open_, level)
                shares = self.lot_cash / price
                self.lots.append((price, shares))
                self.invested += self.lot_cash
                self.fills.append({"i": i, "a": "buy", "p": round(price, 4)})
                self.buy_ptr += 1
                self.consec_buy += 1
                self.consec_sell = 0
                if self.consec_buy >= MAX_CONSECUTIVE:
                    self.closed = True
            else:
                break
        # 卖出：high 触及卖档（卖出 FIFO 最早档，基准更新为卖价）
        while self.sell_ptr < len(SELL_LEVELS) and self.lots and not self.closed:
            level = self.base * (1.0 + SELL_LEVELS[self.sell_ptr])
            if high >= level:
                price = max(open_, level)
                buy_price, shares = self.lots.pop(0)
                self.cash_out += price * shares
                self.realized_pnl += (price - buy_price) * shares
                self.fills.append(
                    {"i": i, "a": "sell", "p": round(price, 4), "bp": round(buy_price, 4)}
                )
                self.sell_ptr += 1
                self.consec_sell += 1
                self.consec_buy = 0
                self.base = price          # 基准价更新，档位重置
                self.buy_ptr = 0
                self.sell_ptr = 0
                if self.consec_sell >= MAX_CONSECUTIVE:
                    self.closed = True
            else:
                break

    def market_value(self, close: float) -> float:
        """剩余持仓按收盘价估值。"""
        return sum(shares * close for _, shares in self.lots)

    def force_close(self, close: float, i: int = -1) -> float:
        """保护触发后的清仓：返回实现总值，持仓清空。"""
        value = self.market_value(close)
        for buy_price, shares in self.lots:
            self.realized_pnl += (close - buy_price) * shares
            self.fills.append(
                {"i": i, "a": "close", "p": round(close, 4), "bp": round(buy_price, 4)}
            )
        self.cash_out += value
        self.lots = []
        self.closed = True
        return value


def simple_grid_backtest(df) -> dict:
    """
    纯网格对照：无视市场状态连续运行网格，保护关闭即重启。

    展示"网格死于单边"的对照——趋势段连环触发买入档，自我保护
    反复清仓实现亏损。
    """
    opens = df["open"].astype(float).tolist()
    highs = df["high"].astype(float).tolist()
    lows = df["low"].astype(float).tolist()
    closes = df["close"].astype(float).tolist()

    pool = 1.0
    grid = None
    rounds = 0
    equity_curve = [1.0]
    for i in range(len(df)):
        if grid is None:
            grid = GridEngine(opens[i], pool / 8.0)
            rounds += 1
        grid.on_day(i, opens[i], highs[i], lows[i])
        if grid.closed:
            grid.force_close(closes[i], i)
            pool = pool - grid.invested + grid.cash_out
            grid = None
            equity = pool
        else:
            equity = (
                pool - grid.invested + grid.cash_out + grid.market_value(closes[i])
            )
        equity_curve.append(equity)
    if grid is not None:
        grid.force_close(closes[-1], len(df) - 1)
        pool = pool - grid.invested + grid.cash_out
        equity_curve.append(pool)

    eq = pd.Series(equity_curve)
    return {
        "total_return": pool - 1.0,
        "max_dd": float((eq / eq.cummax() - 1.0).min()),
        "rounds": rounds,
    }


def switch_backtest(df, hold, regime) -> dict:
    """
    复合系统切换回测：市场状态裁决策略。

    - hold 多头（昨日判定）      -> 趋势层：单笔管理（转多次日开盘
      全仓买、转空次日开盘卖）
    - hold 空头 & regime（昨日判定）-> 网格层：开盘价启动，每档 1/8 资金
    - 其余                        -> 现金（观望）
    - 趋势优先：hold 转多时若网格在跑，先按开盘价平网格再进趋势
    - 网格自我保护关闭 -> 清仓转观望，regime 再次满足可重启

    regime 为"当日判定"的震荡掩码（三条件全过），内部统一 shift(1)
    保证无未来函数。
    """
    opens = df["open"].astype(float).tolist()
    highs = df["high"].astype(float).tolist()
    lows = df["low"].astype(float).tolist()
    closes = df["close"].astype(float).tolist()
    n = len(df)
    prev_hold = hold.shift(1).fillna(False).astype(bool).tolist()
    prev_regime = regime.shift(1).fillna(False).astype(bool).tolist()

    equity = 1.0
    base = 1.0             # 趋势段起点净值
    trend_entry = 0.0
    trend_holding = False
    trend_trades: list[dict] = []
    grid = None
    grid_pool = 0.0
    grid_rounds = 0
    grid_round_list: list[dict] = []
    grid_start_i = -1
    mode_days = {"trend": 0, "grid": 0, "cash": 0}
    mode_series: list[str] = []
    equity_curve: list[float] = [1.0]

    def settle_grid(i: int, price: float) -> None:
        """结束一轮网格：记录轮明细并回收资金。"""
        nonlocal equity, grid, grid_pool
        grid.force_close(price, i)
        ret = (grid.cash_out - grid.invested + grid.market_value(price)) / grid_pool
        grid_round_list.append(
            {
                "si": grid_start_i, "ei": i,
                "ret": round((ret - 1.0) * 100, 2),
                "fills": grid.fills,
            }
        )
        equity = grid_pool - grid.invested + grid.cash_out
        grid = None

    for i in range(n):
        # ---- 开盘：趋势层调仓（优先级最高）----
        if prev_hold[i] and not trend_holding:
            if grid is not None:  # 趋势优先，平掉网格
                settle_grid(i, opens[i])
            base = equity
            trend_entry = opens[i]
            trend_entry_i = i
            trend_holding = True
        elif not prev_hold[i] and trend_holding:
            equity = base * opens[i] / trend_entry
            trend_trades.append(
                {
                    "bi": trend_entry_i, "si": i,
                    "ret": round((opens[i] / trend_entry - 1.0) * 100, 2),
                    "days": i - trend_entry_i,
                }
            )
            trend_holding = False
            base = equity
        # ---- 网格启动（昨日判定：趋势空仓 + 震荡验证通过）----
        if (
            not trend_holding
            and grid is None
            and not prev_hold[i]
            and prev_regime[i]
        ):
            grid_pool = equity
            grid_start_i = i
            grid = GridEngine(opens[i], grid_pool / 8.0)
            grid_rounds += 1
        # ---- 网格运行 ----
        if grid is not None:
            grid.on_day(i, opens[i], highs[i], lows[i])
            if grid.closed:
                settle_grid(i, closes[i])
        # ---- 净值与模式统计 ----
        if trend_holding:
            equity = base * closes[i] / trend_entry
            mode_days["trend"] += 1
            mode_series.append("trend")
        elif grid is not None:
            equity = (
                grid_pool - grid.invested + grid.cash_out + grid.market_value(closes[i])
            )
            mode_days["grid"] += 1
            mode_series.append("grid")
        else:
            mode_days["cash"] += 1
            mode_series.append("cash")
        equity_curve.append(equity)

    # 期末平仓
    if trend_holding:
        equity = base * closes[-1] / trend_entry
        trend_trades.append(
            {
                "bi": trend_entry_i, "si": n - 1,
                "ret": round((closes[-1] / trend_entry - 1.0) * 100, 2),
                "days": n - 1 - trend_entry_i,
            }
        )
        equity_curve.append(equity)
    if grid is not None:
        settle_grid(n - 1, closes[-1])
        equity_curve.append(equity)

    eq = pd.Series(equity_curve)
    rets = [t["ret"] / 100 for t in trend_trades]
    return {
        "total_return": equity - 1.0,
        "max_dd": float((eq / eq.cummax() - 1.0).min()),
        "trend_trades": len(trend_trades),
        "trend_win_rate": (
            sum(1 for r in rets if r > 0) / len(rets) if rets else None
        ),
        "grid_rounds": grid_rounds,
        "mode_days": mode_days,
        "equity_curve": eq,
        "mode_series": mode_series,
        "trend_trade_list": trend_trades,
        "grid_round_list": grid_round_list,
    }
