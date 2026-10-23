"""金叉死叉分析：读取 fetch_data.py 缓存的一体化 CSV，输出文本报告。

不联网、不重复计算指标——CSV 里的 ma_signal / macd_signal 列已由
indicators.py 算好，本脚本只做三类事：

1. 当前状态：最新K线的 MA/MACD 多空排列、处于金叉还是死叉状态
2. 信号统计：数量、频率、MA 与 MACD 的灵敏度对比、最近信号明细
3. 简单回测：金叉次日开盘买入、死叉次日开盘卖出，对比买入持有

用法（在本目录下）：

    python3 analyze.py                          # 分析 562500 全部已缓存周期
    python3 analyze.py --period daily           # 只看日线
    python3 analyze.py --period 1 --limit 5     # 1 分钟线，最近信号只列 5 条

数据不存在时请先运行 fetch_data.py 拉取。
"""

from __future__ import annotations

import argparse
import json
import re
from collections import Counter
from itertools import combinations
from pathlib import Path

import pandas as pd

from grid import simple_grid_backtest, switch_backtest
from ladder import ladder_backtest

DATA_DIR = Path(__file__).resolve().parent / "data"

META_PATH = DATA_DIR / "meta.json"


def symbol_name(symbol: str) -> str | None:
    """读取证券名称（离线，读 data/meta.json；缺失返回 None 降级只显示代码）。

    名称由 fetch_data.py 拉数时自动缓存，或用 --name 手动补全。
    """
    if not META_PATH.exists():
        return None
    try:
        meta = json.loads(META_PATH.read_text(encoding="utf-8"))
    except (ValueError, OSError):
        return None
    # 先查普通键（股票/ETF），缺失再兜底查指数键（idx: 前缀）
    return meta.get(symbol) or meta.get(f"idx:{symbol}")


def symbol_label(symbol: str) -> str:
    """带名称的证券显示名（"562500 芯片ETF"），名称缺失时退化为纯代码。"""
    name = symbol_name(symbol)
    return f"{symbol} {name}" if name else symbol


PERIOD_NAMES = {
    "daily": "日线",
    "weekly": "周线",
    "monthly": "月线",
    "60": "60分钟线",
    "30": "30分钟线",
    "15": "15分钟线",
    "5": "5分钟线",
    "1": "1分钟线",
}

# 分析报告的长周期在前
PERIOD_ORDER = {
    "monthly": 0, "weekly": 1, "daily": 2,
    "60": 3, "30": 4, "15": 5, "5": 6, "1": 7,
}

# 周期许可的高周期映射（低周期信号受高周期状态许可）
PERIOD_HIGHER = {"daily": "weekly"}

# 信号列名约定：ma_{a}x{b}_signal（均线对）与 macd_signal
MA_SIGNAL_RE = re.compile(r"ma_(\d+)x(\d+)_signal")


def ma_columns(df: pd.DataFrame) -> list[str]:
    """探测均线数值列（ma5/ma10/ma20...），按窗口升序。

    fetch_data.py 的 --ma 参数决定列名（如 --ma 3,10 生成 ma3/ma10），
    因此这里动态探测而不是写死。
    """
    return sorted(
        (c for c in df.columns if c.startswith("ma") and c[2:].isdigit()),
        key=lambda c: int(c[2:]),
    )


def signal_names(df: pd.DataFrame) -> dict[str, str]:
    """信号列 -> 显示名，均线对按窗口从小到大（短/中/长）排列，MACD 最后。

    列名约定（fetch_data.py 生成）：ma_{a}x{b}_signal，如 ma_5x10_signal。
    兼容旧缓存的单列 ma_signal（按首尾均线命名，排在最前）。
    """
    names: dict[str, str] = {}
    pairs = []
    for col in df.columns:
        m = MA_SIGNAL_RE.fullmatch(col)
        if m:
            pairs.append((int(m.group(1)), int(m.group(2)), col))
    for a, b, col in sorted(pairs):
        names[col] = f"MA{a}x{b}"
    if "ma_signal" in df.columns and not pairs:
        cols = ma_columns(df)
        if len(cols) >= 2:
            legacy = {"ma_signal": f"MA{cols[0][2:]}x{cols[-1][2:]}"}
            names = {**legacy, **names}
    if "macd_signal" in df.columns:
        names["macd_signal"] = "MACD"
    return names


def load(symbol: str, period: str) -> pd.DataFrame:
    """读取一体化缓存 CSV。"""
    path = DATA_DIR / f"{symbol}_{period}.csv"
    if not path.exists():
        raise FileNotFoundError(
            f"缺少 {path}，请先运行: "
            f"python3 fetch_data.py --symbol {symbol} --period {period}"
        )
    df = pd.read_csv(path)
    df["date"] = pd.to_datetime(df["date"])
    return df.reset_index(drop=True)


def fmt_pct(x: float) -> str:
    return f"{x * 100:+.1f}%"


def regime_report(df: pd.DataFrame) -> None:
    """
    市场状态判定：趋势市 / 震荡市 / 过渡状态。

    对照复合系统"第二道门"的三条震荡判定标准（纯指标层面；
    第一道门"趋势系统空闲"属于持仓状态机，不在本报告范围）：
      ① ADX < 25 且近 5 根未持续上行（排除强趋势加速）
      ② 多均线粘合度 < 3%（排除发散行情）
      ③ 价格在 Donchian 区间内运行（排除突破新高/新低的破位行情）
    三条件全满足 = 震荡市（可开网格）；0/3 或 ADX >= 35 = 趋势市。
    """
    required = {"adx", "pdi", "mdi", "cohesion", "don_upper", "don_lower"}
    if not required.issubset(df.columns):
        print("[市场状态判定] 缓存缺少 regime 列，请 --force 重拉数据")
        return

    last = df.iloc[-1]
    close = float(last["close"])
    adx_val = float(last["adx"])
    pdi_v, mdi_v = float(last["pdi"]), float(last["mdi"])
    cohesion = float(last["cohesion"])
    don_up, don_low = float(last["don_upper"]), float(last["don_lower"])

    adx_recent = df["adx"].tail(6)
    adx_rising = bool(adx_recent.iloc[-1] > adx_recent.iloc[0])
    cond_adx = (adx_val < 25.0) and (not adx_rising)

    cond_ma = cohesion < 0.03

    in_range = don_low <= close <= don_up
    if don_up > don_low:
        pos = (close - don_low) / (don_up - don_low)
        pos_text = f"区间内位置 {pos:.0%}"
    else:
        pos_text = "区间退化（高低同点）"

    print("[市场状态判定] 震荡三条件（全满足=震荡市，可开网格）")
    checks = (
        ("① 趋势强度", f"ADX={adx_val:.1f}"
                       f"（{'<25' if adx_val < 25 else '>=25'}，"
                       f"近5根{'上行' if adx_rising else '未上行'}）", cond_adx),
        ("② 均线形态", f"粘合度={cohesion:.1%}"
                       f"（{'<3% 粘合' if cond_ma else '>=3% 发散'}）", cond_ma),
        ("③ 价格位置", f"close={close:g}，"
                       f"区间[{don_low:g}, {don_up:g}]，{pos_text}", in_range),
    )
    for label, detail, ok in checks:
        print(f"  {label}: {detail} -> {'满足' if ok else '不满足'}")

    score = sum(1 for _, _, ok in checks if ok)
    if score == 3:
        verdict = "震荡市（第二道门通过，可考虑网格）"
    elif score == 0 or adx_val >= 35.0:
        direction = "上行" if pdi_v > mdi_v else "下行"
        verdict = f"趋势市（{direction}趋势）"
    else:
        verdict = "过渡状态（趋势与震荡拉锯，网格与趋势都不宜重仓）"
    print(f"  判定: {score}/3 -> {verdict}")

    if pdi_v > mdi_v:
        direction = "上行动能占优"
    elif mdi_v > pdi_v:
        direction = "下行动能占优"
    else:
        direction = "多空均衡"
    print(f"  参考: +DI({pdi_v:.1f}) vs -DI({mdi_v:.1f}) -> {direction}")


def snapshot(df: pd.DataFrame) -> None:
    """打印最新K线各信号的多空状态（均线对按短/中/长排列）。"""
    last = df.iloc[-1]
    date = last["date"].strftime("%Y-%m-%d")
    print(f"[当前状态] {date}  close={last['close']:g}")
    rows = []
    for col, name in signal_names(df).items():
        if col == "macd_signal":
            rows.append((col, name, "DIF", last["dif"], "DEA", last["dea"]))
            continue
        m = MA_SIGNAL_RE.fullmatch(col)
        if m:
            a, b = m.group(1), m.group(2)
            rows.append(
                (col, name, f"MA{a}", last[f"ma{a}"], f"MA{b}", last[f"ma{b}"])
            )
        else:  # 旧缓存单列 ma_signal：取首尾两条均线
            cols = ma_columns(df)
            if len(cols) >= 2:
                rows.append(
                    (col, name, cols[0].upper(), last[cols[0]],
                     cols[-1].upper(), last[cols[-1]])
                )
    for col, name, fast_label, fast_v, slow_label, slow_v in rows:
        fast, slow = float(fast_v), float(slow_v)
        bull = fast > slow
        sig_rows = df[df[col].notna()]
        since = (
            sig_rows.iloc[-1]["date"].strftime("%Y-%m-%d")
            if not sig_rows.empty
            else "无"
        )
        state = "金叉状态(多头)" if bull else "死叉状态(空头)"
        arrow = ">" if bull else "<"
        print(
            f"  {name}: {fast_label}({fast:.4g}) {arrow} {slow_label}({slow:.4g})"
            f" -> {state} (自 {since})"
        )
    # 背离预警：最近 10 根内出现过背离则提示（预警而非操作信号）
    if {"bear_div", "bull_div"}.issubset(df.columns):
        recent = df.tail(10)
        if recent["bear_div"].any():
            print("  预警: 最近10根出现顶背离（价格新高而动能未新高）")
        if recent["bull_div"].any():
            print("  预警: 最近10根出现底背离（价格新低而动能未新低）")


def signal_stats(df: pd.DataFrame) -> None:
    """信号数量、频率与两套指标的灵敏度对比。

    频率口径随数据跨度自适应：>= 1 年按"年均"，不足 1 年按"日均"
    （分钟线只有几个交易日，按年折算出的数字没有意义）。
    """
    days = max((df["date"].iloc[-1] - df["date"].iloc[0]).days, 1)
    names = signal_names(df)
    if days >= 365:
        span_text = f"跨度约 {days / 365.25:.1f} 年"
        freq_label, freq_div = "年均", days / 365.25
    else:
        span_text = f"跨度 {days} 天"
        freq_label, freq_div = "日均", days
    print(f"[信号统计] {len(df)} 根K线, {span_text}")
    counts: dict[str, int] = {}
    for col, name in names.items():
        golden = int((df[col] == "golden").sum())
        death = int((df[col] == "death").sum())
        total = golden + death
        counts[col] = total
        if total:
            freq = f"{total / freq_div:.1f}"
            interval = f"{len(df) / total:.1f}"
        else:
            freq = interval = "--"
        print(
            f"  {name:<8} | 金叉 {golden} | 死叉 {death} | 合计 {total}"
            f" | {freq_label} {freq} | 平均间隔 {interval} 根K线"
        )
    ma_cols = [c for c in names if c.startswith("ma_")]
    if len(ma_cols) >= 2 and counts.get(ma_cols[0]) and counts.get(ma_cols[-1]):
        fastest, slowest = ma_cols[0], ma_cols[-1]
        print(
            f"  灵敏度: {names[fastest]} 触发频率是 {names[slowest]} 的"
            f" {counts[fastest] / counts[slowest]:.2f} 倍"
        )


def recent_signals(df: pd.DataFrame, limit: int) -> None:
    """两套指标的信号合并按时间排序，展示最近 limit 个。"""
    frames = []
    for col, name in signal_names(df).items():
        sub = df[df[col].notna()][["date", "close", col]].copy()
        if sub.empty:
            continue
        sub = sub.rename(columns={col: "kind"})
        sub["source"] = name
        frames.append(sub)
    print(f"[最近信号] 最近 {limit} 个（两套指标合并按时间排序）")
    if not frames:
        print("  无任何信号")
        return
    merged = (
        pd.concat(frames, ignore_index=True).sort_values("date").tail(limit)
    )
    # 日线/周线的时间部分全为 00:00，只显示日期；分钟线保留时分
    has_time = ((df["date"].dt.hour != 0) | (df["date"].dt.minute != 0)).any()
    fmt = "%Y-%m-%d %H:%M" if has_time else "%Y-%m-%d"
    for _, row in merged.iterrows():
        kind = "金叉" if row["kind"] == "golden" else "死叉"
        print(
            f"  {row['date'].strftime(fmt)}"
            f"  {kind}  close={row['close']:g}  ({row['source']})"
        )


def buy_and_hold(df: pd.DataFrame) -> tuple[float, float]:
    """买入持有基准：首日开盘买入持有到期末收盘，返回 (收益, 最大回撤)。"""
    closes = df["close"].astype(float)
    equity = closes / float(df["open"].iloc[0])
    peak = equity.cummax()
    return float(equity.iloc[-1] - 1.0), float((equity / peak - 1.0).min())


def state_backtest(df: pd.DataFrame, hold: pd.Series) -> dict:
    """
    状态驱动回测：hold 为 True 的日子收盘后应持仓，次日开盘调仓。

    与事件驱动 backtest() 同一成交规则（次日开盘、期末按收盘平仓、
    不计手续费）。对单指标而言两个引擎应给出一致结果（金叉事件
    = 多头状态从无到有，持有区间相同），互为交叉验证。两个已知
    语义边界（非 bug）：
    - 状态版若数据开头即处于多头（无金叉事件）也会入场；
    - 均线恰好相等的日子（diff=0）：事件版不算穿越、晚一日离场；
      状态版严格比较（>）、当日即失去多头资格，提前离场。
      （562500 日线 5 年中仅出现一次，实测影响 0.8% 收益差）

    净值曲线与最大回撤的计算同 backtest()：日收益连乘
    （买入日 close/open、持有日 close/昨收、卖出日 open/昨收）。
    """
    should = hold.shift(1).fillna(False).tolist()
    opens = df["open"].astype(float).tolist()
    closes = df["close"].astype(float).tolist()
    trades: list[tuple[float, float]] = []
    holding: float | None = None
    equity = 1.0
    peak = 1.0
    max_dd = 0.0
    equity_curve: list[float] = [1.0]

    for i in range(len(df)):
        if should[i] and holding is None:
            holding = opens[i]
            day_ret = closes[i] / opens[i]
        elif not should[i] and holding is not None:
            trades.append((holding, opens[i]))
            day_ret = opens[i] / closes[i - 1]
            holding = None
        elif holding is not None:
            day_ret = closes[i] / closes[i - 1]
        else:
            day_ret = 1.0
        if day_ret != 1.0:
            equity *= day_ret
            if equity > peak:
                peak = equity
            dd = equity / peak - 1.0
            if dd < max_dd:
                max_dd = dd
        equity_curve.append(equity)

    closed_at_end = False
    if holding is not None:
        trades.append((holding, closes[-1]))
        closed_at_end = True

    total = 1.0
    for buy, sell in trades:
        total *= sell / buy
    total -= 1.0
    wins = sum(1 for buy, sell in trades if sell > buy)
    return {
        "trades": len(trades),
        "win_rate": (wins / len(trades)) if trades else None,
        "total_return": total,
        "max_dd": max_dd,
        "closed_at_end": closed_at_end,
        "equity_curve": pd.Series(equity_curve),
    }


def signal_bulls(df: pd.DataFrame) -> dict[str, pd.Series]:
    """各信号源的多头状态掩码。

    均线对 MA(a)x(b) 多头 = 当日 MA(a) > MA(b)；MACD 多头 = DIF > DEA。
    返回 {显示名: 掩码}，顺序与 signal_names 一致（均线按窗口短->长，
    MACD 最后）。
    """
    bulls: dict[str, pd.Series] = {}
    for col, name in signal_names(df).items():
        if col == "macd_signal":
            bulls[name] = df["dif"] > df["dea"]
            continue
        m = MA_SIGNAL_RE.fullmatch(col)
        if m:
            bulls[name] = df[f"ma{m.group(1)}"] > df[f"ma{m.group(2)}"]
    return bulls


def _has_transitive_redundancy(combo: tuple[str, ...]) -> bool:
    """
    检测均线链传递性冗余：组合同时含 MA(a)x(b) 与 MA(b)x(c) 时，
    MA(a)x(c) 被前两者数学蕴含（a>b 且 b>c 推出 a>c），属于冗余成员。
    含冗余成员的组合与剔除冗余后的更小组合完全等价，无需重复回测。
    """
    pairs = []
    for name in combo:
        m = re.fullmatch(r"MA(\d+)x(\d+)", name)
        if m:
            pairs.append((int(m.group(1)), int(m.group(2))))
    pair_set = set(pairs)
    for a, b in pairs:
        for c, d in pairs:
            if b == c and (a, d) in pair_set:
                return True
    return False


def build_filters(
    df: pd.DataFrame, higher_df: pd.DataFrame | None, pair: str,
) -> list[tuple[str, pd.Series]]:
    """构建 6 种环境过滤器（ADX/DIF/底背离/周线许可），返回 [(名称, 掩码)]。

    各报告段、密度策略、图表 payload 共用此函数，保证口径一致。
    """
    filters: list[tuple[str, pd.Series]] = []
    if "adx" in df.columns:
        filters.append(("ADX>=20", df["adx"] >= 20))
        filters.append(("ADX>=25", df["adx"] >= 25))
    if "dif_pos" in df.columns:
        filters.append(("DIF零上", df["dif_pos"]))
    if "bull_div" in df.columns:
        filters.append((
            "底背离前兆",
            df["bull_div"].astype(bool).rolling(10, min_periods=1).max().astype(bool),
        ))
    if higher_df is not None:
        for src, lab in ((pair, f"周线{pair}"), ("MACD", "周线MACD")):
            mask = higher_permission(df, higher_df, src)
            if mask is not None:
                filters.append((f"{lab}多头", mask))
    return filters


def build_all_masks(
    df: pd.DataFrame, bulls: dict[str, pd.Series],
    combo: pd.Series, higher_df: pd.DataFrame | None,
) -> list[tuple[str, pd.Series]]:
    """构建全部掩码型策略的 (名称, 持仓掩码) 列表。

    包含：单指标 + 双确认 + 三确认（跳过传递性冗余）+ OR 对偶 + 过滤版。
    signal_density_report、density_hold、compute_density 共用此函数。
    """
    names = list(bulls)
    masks: list[tuple[str, pd.Series]] = []
    # 单指标
    for name, hold in bulls.items():
        masks.append((name, hold))
    # 双确认 / 三确认
    for size in (2, 3):
        for c in combinations(names, size):
            if size == 3 and _has_transitive_redundancy(c):
                continue
            n = "+".join(c)
            h = bulls[c[0]]
            for m in c[1:]:
                h = h & bulls[m]
            masks.append((n, h))
    # OR 对偶
    or_hold = None
    for h in bulls.values():
        or_hold = h if or_hold is None else or_hold | h
    masks.append(("任一确认(OR)", or_hold))
    # 过滤版
    pair = None
    cols = ma_columns(df)
    if cols and "MACD" in bulls:
        pair = f"MA{cols[0][2:]}x{cols[-1][2:]}"
    if pair and combo is not None:
        for label, fmask in build_filters(df, higher_df, pair):
            masks.append((f"{pair}+MACD+{label}", combo & fmask))
    return masks


def collect_density_triggers(
    df: pd.DataFrame, bulls: dict[str, pd.Series], combo: pd.Series,
    rmask: pd.Series | None, higher_df: pd.DataFrame | None,
    n_days: int,
) -> list[tuple[str, str, pd.Timestamp]]:
    """收集最近 N 天内全部策略的信号触发事件。

    返回 [(策略名, 方向, 日期)]，方向为 "买入"/"卖出"/"预警"。
    signal_density_report 和 chart.compute_density 共用此函数。
    """
    dates = df["date"]
    last_date = dates.iloc[-1]
    cutoff = last_date - pd.Timedelta(days=n_days)
    recent = dates >= cutoff

    triggers: list[tuple[str, str, pd.Timestamp]] = []

    def add_mask(name: str, hold: pd.Series) -> None:
        prev = hold.shift(1, fill_value=hold.iloc[0])
        for d in dates[(~prev.astype(bool)) & hold & recent]:
            triggers.append((name, "买入", d))
        for d in dates[prev.astype(bool) & (~hold) & recent]:
            triggers.append((name, "卖出", d))

    # 掩码型策略（单指标 + 双确认 + 三确认 + OR + 过滤版）
    for name, hold in build_all_masks(df, bulls, combo, higher_df):
        add_mask(name, hold)

    # 三档阶梯止盈
    r = ladder_backtest(df, combo, exit_on_signal=True, cooldown=True)
    for t in r["trades"]:
        entry_d = dates.iloc[t["entry_i"]]
        if entry_d >= cutoff:
            triggers.append(("三档+冷却协议", "买入", entry_d))
        last_exit_i = max(i for i in t["exit_is"] if i is not None)
        exit_d = dates.iloc[last_exit_i]
        if exit_d >= cutoff:
            triggers.append(("三档+冷却协议", "卖出", exit_d))

    # 切换系统
    if rmask is not None:
        r = switch_backtest(df, combo, rmask)
        for t in r["trend_trade_list"]:
            if dates.iloc[t["bi"]] >= cutoff:
                triggers.append(("切换系统", "买入", dates.iloc[t["bi"]]))
            if dates.iloc[t["si"]] >= cutoff:
                triggers.append(("切换系统", "卖出", dates.iloc[t["si"]]))
        if higher_df is not None:
            wperm = higher_permission(df, higher_df, "MACD")
            if wperm is not None:
                r2 = switch_backtest(df, combo, rmask & wperm)
                for t in r2["trend_trade_list"]:
                    if dates.iloc[t["bi"]] >= cutoff:
                        triggers.append(("切换+网格方向过滤", "买入", dates.iloc[t["bi"]]))
                    if dates.iloc[t["si"]] >= cutoff:
                        triggers.append(("切换+网格方向过滤", "卖出", dates.iloc[t["si"]]))

    # 背离预警
    if {"bull_div", "bear_div"}.issubset(df.columns):
        for d in dates[recent & df["bull_div"].fillna(False).astype(bool)]:
            triggers.append(("底背离(预警)", "预警", d))
        for d in dates[recent & df["bear_div"].fillna(False).astype(bool)]:
            triggers.append(("顶背离(预警)", "预警", d))

    return triggers


def multi_combo_report(df: pd.DataFrame) -> None:
    """
    多信号确认回测：从 N=1（单指标）到 N=4（四重确认）逐级推广。

    组合语义（每种组合是一条独立策略）：
    - 持仓条件：组合成员在昨日收盘时全部处于多头状态（AND 逻辑）
    - 调仓执行：今日开盘按上述判定建仓/清仓（无未来函数）
    - 多头判定：均线对 MA(a)x(b) 多头 = 当日 MA(a) > MA(b)；
      MACD 多头 = 当日 DIF > DEA
    - 期末持仓按收盘平仓，不计手续费

    末尾附 AND 的对偶对照（任一确认 OR：任一成员多头即持仓）。
    均线链传递性冗余的组合跳过不回测，单独列出。
    """
    bulls = signal_bulls(df)
    if len(bulls) < 2:
        print("[多信号确认] 信号源不足，跳过")
        return

    # 每个信号源的多头判定公式——说清楚"组合"到底怎么执行
    judges = "; ".join(
        "MACD多头=DIF>DEA"
        if n == "MACD"
        else f"{n}多头=" + re.sub(r"MA(\d+)x(\d+)", r"MA\1>MA\2", n)
        for n in bulls
    )
    print(
        f"[多信号确认] {len(bulls)} 个信号源，N=1 到 N={len(bulls)} 逐级推广"
    )
    print(
        "  组合语义: 成员全部多头才持仓(AND)"
        " | 昨日收盘判定, 次日开盘调仓 | 不计手续费"
    )
    print(f"  多头判定: {judges}")

    bh, bh_dd = buy_and_hold(df)
    print(f"  基准: 买入持有 {fmt_pct(bh)}（回撤 {fmt_pct(bh_dd)}），超额 = 收益 - 持有")
    skipped: list[str] = []
    best_name, best_ret, best_dd, best_level = "", -float("inf"), 0.0, ""
    level_names = {1: "单指标", 2: "双确认", 3: "三确认", 4: "四确认"}

    for size in range(1, len(bulls) + 1):
        level = level_names.get(size, f"{size}确认")
        rows = []
        for combo in combinations(bulls, size):
            if _has_transitive_redundancy(combo):
                skipped.append("+".join(combo))
                continue
            hold = None
            for member in combo:
                hold = bulls[member] if hold is None else hold & bulls[member]
            rows.append((combo, state_backtest(df, hold)))
        if not rows:
            print(f"  N={size} {level}: 全部为传递性冗余组合，无有效策略")
            continue
        rows.sort(key=lambda x: x[1]["total_return"], reverse=True)
        print(f"  N={size} {level}（{len(rows)} 种，按收益排序）:")
        for combo, r in rows:
            wr = f"{r['win_rate'] * 100:.1f}%" if r["win_rate"] is not None else "--"
            combo_name = "+".join(combo)
            if r["total_return"] > best_ret:
                best_name, best_ret, best_dd = combo_name, r["total_return"], r["max_dd"]
                best_level = level
            print(
                f"    {combo_name:<28} | {r['trades']:>3} 笔 | 胜率 {wr:>6}"
                f" | 收益 {fmt_pct(r['total_return']):>7}"
                f" | 回撤 {fmt_pct(r['max_dd']):>7}"
                f" | 超额 {fmt_pct(r['total_return'] - bh):>7}"
            )

    # AND 的对偶对照：任一成员多头即持仓（宽松光谱端）
    any_hold = None
    for hold in bulls.values():
        any_hold = hold if any_hold is None else any_hold | hold
    r = state_backtest(df, any_hold)
    wr = f"{r['win_rate'] * 100:.1f}%" if r["win_rate"] is not None else "--"
    print(
        f"  对偶对照: 任一确认(OR，{len(bulls)} 个源任一多头即持仓)"
        f" | {r['trades']:>3} 笔 | 胜率 {wr:>6}"
        f" | 收益 {fmt_pct(r['total_return']):>7}"
        f" | 回撤 {fmt_pct(r['max_dd']):>7}"
        f" | 超额 {fmt_pct(r['total_return'] - bh):>7}"
    )
    print(
        f"  冠军: {best_name}（{best_level}，收益 {fmt_pct(best_ret)}，"
        f"回撤 {fmt_pct(best_dd)}，超额 {fmt_pct(best_ret - bh)}）"
    )
    if skipped:
        print(
            f"  冗余跳过 {len(skipped)} 种: {'、'.join(skipped)}"
            f"（均线链传递性等价于已列组合，不重复回测）"
        )


def higher_permission(
    day_df: pd.DataFrame, higher_df: pd.DataFrame, source: str
) -> pd.Series | None:
    """
    周期许可掩码：高周期信号源的多头状态对齐到低周期每个交易日。

    对齐规则（无未来函数）：交易日 t 的收盘判定可用 date <= t 的
    最新高周期状态——date == t 的高周期K线恰在 t 收盘成型，配合
    state_backtest"昨日收盘判定、次日开盘执行"的语义合法。
    高周期数据起点之前的交易日视为无许可（False）。
    """
    hb = signal_bulls(higher_df)
    if source not in hb:
        return None
    state = higher_df[["date"]].copy()
    state["bull"] = hb[source].to_numpy()
    merged = pd.merge_asof(
        day_df[["date"]], state, on="date", direction="backward"
    )
    return pd.Series(merged["bull"].fillna(False).to_numpy(), index=day_df.index)


def filter_report(df: pd.DataFrame, higher_df: pd.DataFrame | None = None) -> None:
    """
    条件过滤回测：信号源 × 环境条件，量化"过滤值多少超额"。

    持仓条件 = 信号多头 AND 过滤器——环境不满足时不开火；注意这是
    状态语义：持仓期间过滤器失效（如 ADX 跌破阈值）也会清仓，比
    "只过滤入场"更激进。每行超额 = vs 同信号源的原始（无过滤）策略。

    过滤器族：
    - 趋势强度：ADX >= 20 / 25（单周期内，无趋势不做趋势信号）
    - 周期许可：高周期（如周线）MA / MACD 多头才允许低周期信号开火
      （高周期状态按 date<=t 对齐，无未来函数；需提供 higher_df）
    """
    if "adx" not in df.columns:
        print("[条件过滤] 缓存缺少 adx 列（--recompute 刷新），跳过")
        return
    bulls = signal_bulls(df)
    cols = ma_columns(df)
    # 双确认代表：首尾均线对 × MACD（与多信号段的默认组合一致）
    if len(cols) >= 2 and "MACD" in bulls:
        pair = f"MA{cols[0][2:]}x{cols[-1][2:]}"
        if pair in bulls:
            bulls[f"{pair}+MACD"] = bulls[pair] & bulls["MACD"]
    if len(bulls) < 2:
        print("[条件过滤] 信号源不足，跳过")
        return

    bh, _ = buy_and_hold(df)
    filters: list[tuple[str, pd.Series | None]] = [
        ("原始(无过滤)", None),
        ("ADX>=20", df["adx"] >= 20),
        ("ADX>=25", df["adx"] >= 25),
    ]
    if "dif_pos" in df.columns:
        filters.append(("DIF零上", df["dif_pos"]))
    if "bull_div" in df.columns:
        # 前 10 根内出现过底背离才允许开火（标签分析 5/5 源有信息量后升级）
        recent_div = (
            df["bull_div"].astype(bool).rolling(10, min_periods=1).max().astype(bool)
        )
        filters.append(("底背离前兆(10根)", recent_div))
    if higher_df is not None:
        pair = f"MA{cols[0][2:]}x{cols[-1][2:]}" if len(cols) >= 2 else None
        for label, source in (
            (f"周线{pair}多头", pair),
            ("周线MACD多头", "MACD"),
        ):
            if source is None:
                continue
            mask = higher_permission(df, higher_df, source)
            if mask is not None:
                filters.append((label, mask))

    # 按原始收益降序排列信号源
    order = sorted(
        bulls, key=lambda n: -state_backtest(df, bulls[n])["total_return"]
    )
    print(
        "[条件过滤] 持仓 = 信号多头 AND 过滤器"
        "（趋势不足/高周期不许可时不开火，持仓中失效即清仓）"
    )
    print(f"  基准: 买入持有 {fmt_pct(bh)} | 每行超额 = vs 同信号原始策略")
    for sname in order:
        bull = bulls[sname]
        base: float | None = None
        print(f"  {sname}:")
        for fname, fmask in filters:
            hold = bull if fmask is None else bull & fmask
            r = state_backtest(df, hold)
            wr = f"{r['win_rate'] * 100:.1f}%" if r["win_rate"] is not None else "--"
            if fmask is None:
                base = r["total_return"]
                excess = "      --"
            else:
                excess = fmt_pct(r["total_return"] - base)
            print(
                f"    {fname:<10} | {r['trades']:>3} 笔 | 胜率 {wr:>6}"
                f" | 收益 {fmt_pct(r['total_return']):>7}"
                f" | 回撤 {fmt_pct(r['max_dd']):>7}"
                f" | 超额(vs原始) {excess:>7}"
            )


def signal_quality_report(
    df: pd.DataFrame,
    forwards: tuple[int, ...] = (10, 20, 30),
    lookback: int = 10,
) -> None:
    """
    信号质量分析：所有信号源（单指标 + 双确认组合）的转多头/转空头
    事件打标签分组，不改变任何规则。

    事件定义：多头掩码的翻转——转多头 = 入场事件（均线金叉或组合
    共振入场），转空头 = 离场事件。

    度量：事件后多尺度收盘收益（默认 10/20/30 根——均线信号平均
    持有约 46 根，单一 10 根尺度只覆盖持有期前段）；入场看涨、离场
    看跌（离场后续为负说明卖对了），正确率按最长尺度计算。
    末端事件的观测窗口不足时按可得长度截断（docstring 诚实记录）。

    三组标签（全部信号源共用，背离/零轴为 MACD 体系概念跨源使用）：
    - 入场 x 零轴位置：事件日 DIF 在零轴上/下
    - 入场 x 底背离前兆：前 lookback 根内出现过底背离
    - 离场 x 顶背离前兆：前 lookback 根内出现过顶背离

    分析先于策略：分组有显著差异才值得升级为过滤规则。
    """
    required = {"dif_pos", "bull_div", "bear_div"}
    if not required.issubset(df.columns):
        print("[信号质量] 缓存缺少背离/零轴列（--recompute 刷新），跳过")
        return
    bulls = signal_bulls(df)
    cols = ma_columns(df)
    if len(cols) >= 2 and "MACD" in bulls:
        pair = f"MA{cols[0][2:]}x{cols[-1][2:]}"
        if pair in bulls:
            bulls[f"{pair}+MACD"] = bulls[pair] & bulls["MACD"]
    if len(bulls) < 2:
        print("[信号质量] 信号源不足，跳过")
        return

    closes = df["close"].astype(float).tolist()
    n = len(df)
    dif_pos = df["dif_pos"].astype(bool)
    bull_div = df["bull_div"].astype(bool)
    bear_div = df["bear_div"].astype(bool)
    main_h = forwards[-1]

    def fwd(i: int, horizon: int) -> float:
        j = min(i + horizon, n - 1)
        return closes[j] / closes[i] - 1.0

    def has_div(div: pd.Series, t: int) -> bool:
        return bool(div.iloc[max(0, t - lookback): t + 1].any())

    def stats(idxs: list[int], good_when_up: bool) -> str:
        if not idxs:
            return "      0 次"
        line = f"{len(idxs):>3} 次 |"
        for h in forwards:
            rets = [fwd(t, h) for t in idxs]
            line += f" {h}根 {sum(rets) / len(rets):+.2%} |"
        main_rets = [fwd(t, main_h) for t in idxs]
        correct = sum(1 for r in main_rets if (r > 0) == good_when_up)
        line += f" {main_h}根正确率 {correct / len(idxs):.1%}"
        return line

    scales = "/".join(str(h) for h in forwards)
    print(
        f"[信号质量] 各信号源转多头/转空头事件 x 背离/零轴"
        f"（事件后 {scales} 根收盘收益，不改变规则）"
    )
    for name, bull in bulls.items():
        prev = bull.shift(1).fillna(False).astype(bool)
        entries = list(df.index[bull & ~prev])
        exits = list(df.index[~bull & prev])
        print(f"  {name}:")
        up = [t for t in entries if dif_pos.iloc[t]]
        down = [t for t in entries if not dif_pos.iloc[t]]
        print(f"    入场 x 零轴   零上: {stats(up, True)}")
        print(f"                  零下: {stats(down, True)}")
        with_div = [t for t in entries if has_div(bull_div, t)]
        without = [t for t in entries if not has_div(bull_div, t)]
        print(f"    入场 x 底背离 有前兆: {stats(with_div, True)}")
        print(f"                  无前兆: {stats(without, True)}")
        w = [t for t in exits if has_div(bear_div, t)]
        wo = [t for t in exits if not has_div(bear_div, t)]
        print(f"    离场 x 顶背离 有前兆: {stats(w, False)}")
        print(f"                  无前兆: {stats(wo, False)}")


def ladder_report(df: pd.DataFrame) -> None:
    """
    三档阶梯止盈 vs 单笔管理：同一入场信号、不同仓位管理的对比。

    信号固定为双确认冠军（首尾均线 x MACD），隔离的唯一变量是
    仓位管理方式——单笔管理（金叉全进死叉全出）对照三档版：
    - 保守版（exit_on_signal=True）：死叉信号清仓全部剩余档
    - 纯移动止损版：第三档完全交给均线梯度，不看死叉
    """
    bulls = signal_bulls(df)
    cols = ma_columns(df)
    if len(cols) < 2 or "MACD" not in bulls:
        print("[三档止盈] 信号源不足，跳过")
        return
    pair = f"MA{cols[0][2:]}x{cols[-1][2:]}"
    hold = bulls[pair] & bulls["MACD"]

    base = state_backtest(df, hold)
    ladder = ladder_backtest(df, hold, exit_on_signal=True)
    pure_trail = ladder_backtest(df, hold, exit_on_signal=False)
    with_cooldown = ladder_backtest(
        df, hold, exit_on_signal=True, cooldown=True
    )

    print(f"[三档止盈] 同信号（{pair}+MACD）不同仓位管理对比:")
    rows = (
        ("单笔管理(金叉进死叉出)", base["trades"], base),
        ("三档+信号清仓(保守)", ladder["count"], ladder),
        ("三档纯移动止损", pure_trail["count"], pure_trail),
        ("三档+假突破冷却协议", with_cooldown["count"], with_cooldown),
    )
    for name, trades_n, r in rows:
        wr = f"{r['win_rate'] * 100:.1f}%" if r["win_rate"] is not None else "--"
        print(
            f"  {name:<20} | {trades_n:>3} 笔 | 胜率 {wr:>6}"
            f" | 收益 {fmt_pct(r['total_return']):>7}"
            f" | 回撤 {fmt_pct(r['max_dd']):>7}"
        )

    reasons = Counter(t["reason"] for t in ladder["trades"])
    t1_hit = sum(1 for t in ladder["trades"] if t["exits"][0] != t["exits"][2])
    t2_hit = sum(1 for t in ladder["trades"] if t["exits"][1] != t["exits"][2])
    print(
        f"  三档细节: 第一档先行触发 {t1_hit} 轮 | 第二档 {t2_hit} 轮"
        f" | 最终离场原因 {dict(reasons)}"
    )


def regime_mask(df: pd.DataFrame) -> pd.Series | None:
    """
    震荡判定掩码（序列版，供切换回测使用）：三条件全过为 True。

    与 regime_report 同一套标准：ADX<25 且近 5 根未上行、均线粘合
    <3%、收盘价在 Donchian 区间内。返回当日判定掩码（调用方需
    shift(1) 保证执行无未来函数）。
    """
    required = {"adx", "cohesion", "don_upper", "don_lower"}
    if not required.issubset(df.columns):
        return None
    adx = df["adx"]
    adx_rising = adx > adx.shift(5)
    cond_adx = (adx < 25.0) & ~adx_rising
    cond_ma = df["cohesion"] < 0.03
    close = df["close"]
    cond_range = (df["don_lower"] <= close) & (close <= df["don_upper"])
    return cond_adx & cond_ma & cond_range


def grid_report(df: pd.DataFrame, higher_df: pd.DataFrame | None = None) -> None:
    """网格与切换系统对比：纯趋势 / 纯网格 / 切换系统 / 买入持有。

    附加变体：网格准入加方向过滤（高周期 MACD 多头才开网格）——
    同一个方向过滤对趋势信号是负贡献（滞后吃趋势初段），对网格
    准入可能是正贡献（拦住下行通道里的网格）。
    """
    bulls = signal_bulls(df)
    cols = ma_columns(df)
    if len(cols) < 2 or "MACD" not in bulls:
        print("[网格与切换] 信号源不足，跳过")
        return
    pair = f"MA{cols[0][2:]}x{cols[-1][2:]}"
    hold = bulls[pair] & bulls["MACD"]
    rmask = regime_mask(df)
    if rmask is None:
        print("[网格与切换] 缓存缺少 regime 列（--recompute 刷新），跳过")
        return

    trend = state_backtest(df, hold)
    pure_grid = simple_grid_backtest(df)
    switched = switch_backtest(df, hold, rmask)
    bh, bh_dd = buy_and_hold(df)

    print(f"[网格与切换] 趋势信号={pair}+MACD | 震荡判定=三条件(ADX/粘合/区间)")
    print(
        f"  纯趋势(双确认)     | {trend['trades']:>3} 笔"
        f" | 收益 {fmt_pct(trend['total_return']):>7}"
        f" | 回撤 {fmt_pct(trend['max_dd']):>7}"
    )
    print(
        f"  纯网格(无视状态)   | {pure_grid['rounds']:>3} 轮"
        f" | 收益 {fmt_pct(pure_grid['total_return']):>7}"
        f" | 回撤 {fmt_pct(pure_grid['max_dd']):>7}"
        f"  <- 趋势段灾难对照"
    )
    print(
        f"  切换系统           | 趋势{switched['trend_trades']}笔/网格{switched['grid_rounds']}轮"
        f" | 收益 {fmt_pct(switched['total_return']):>7}"
        f" | 回撤 {fmt_pct(switched['max_dd']):>7}"
    )
    days = switched["mode_days"]
    total_days = sum(days.values())
    print(
        f"  模式分布: 趋势 {days['trend'] / total_days:.0%}"
        f" | 网格 {days['grid'] / total_days:.0%}"
        f" | 现金 {days['cash'] / total_days:.0%}"
    )
    if higher_df is not None:
        weekly_perm = higher_permission(df, higher_df, "MACD")
        if weekly_perm is not None:
            switched2 = switch_backtest(df, hold, rmask & weekly_perm)
            d2 = switched2["mode_days"]
            t2 = sum(d2.values())
            print(
                f"  切换+网格方向过滤   | 趋势{switched2['trend_trades']}笔"
                f"/网格{switched2['grid_rounds']}轮"
                f" | 收益 {fmt_pct(switched2['total_return']):>7}"
                f" | 回撤 {fmt_pct(switched2['max_dd']):>7}"
            )
            print(
                f"  模式分布: 趋势 {d2['trend'] / t2:.0%}"
                f" | 网格 {d2['grid'] / t2:.0%}"
                f" | 现金 {d2['cash'] / t2:.0%}"
            )
    print(
        f"  买入持有           |   - 笔"
        f" | 收益 {fmt_pct(bh):>7} | 回撤 {fmt_pct(bh_dd):>7}"
    )


def perf_metrics(equity: pd.Series, periods_per_year: int = 252) -> dict:
    """
    绩效指标：总收益 / 年化 / 夏普 / 卡尔玛 / 最大回撤。

    夏普按零无风险利率、等间隔日收益近似；卡尔玛 = 年化收益 /
    |最大回撤|。
    """
    total = float(equity.iloc[-1] / equity.iloc[0] - 1.0)
    n = len(equity) - 1
    annual = 0.0
    if n > 0:
        annual = (1.0 + total) ** (periods_per_year / n) - 1.0
    ret = equity.pct_change().dropna()
    sharpe = None
    if len(ret) > 1 and float(ret.std()) > 0:
        sharpe = float(ret.mean()) / float(ret.std()) * (periods_per_year**0.5)
    dd = float((equity / equity.cummax() - 1.0).min())
    calmar = annual / abs(dd) if dd < 0 else None
    return {
        "total": total,
        "annual": annual,
        "sharpe": sharpe,
        "calmar": calmar,
        "max_dd": dd,
    }


def perf_report(df: pd.DataFrame, higher_df: pd.DataFrame | None, period: str) -> None:
    """
    绩效汇总：关键策略的完整指标对比 + 组合层（底仓 20% + 切换 80%）。

    组合层对应博客三层架构的资金分账简化：底仓永远持有（防踏空），
    共享池跑切换系统。仅支持日线/周线（年化因子不同），分钟线跳过。
    """
    ppy = {"daily": 252, "weekly": 52}.get(period)
    if ppy is None:
        return
    bulls = signal_bulls(df)
    cols = ma_columns(df)
    if len(cols) < 2 or "MACD" not in bulls:
        return
    pair = f"MA{cols[0][2:]}x{cols[-1][2:]}"
    hold = bulls[pair] & bulls["MACD"]
    rmask = regime_mask(df)
    if rmask is None:
        return

    n = len(df)
    closes = df["close"].astype(float)
    hold_curve = pd.Series(
        [1.0] + (closes / float(df["open"].iloc[0])).tolist()
    )
    trend_curve = state_backtest(df, hold)["equity_curve"].iloc[: n + 1]
    ladder_curve = ladder_backtest(
        df, hold, exit_on_signal=True, cooldown=True
    )["equity_curve"].iloc[: n + 1]
    sw = switch_backtest(df, hold, rmask)
    sw_curve = sw["equity_curve"].iloc[: n + 1]
    # 组合层：20% 底仓（永远持有） + 80% 切换系统
    combo_curve = 0.2 * hold_curve + 0.8 * sw_curve

    rows = [
        ("买入持有", hold_curve),
        (f"纯趋势({pair}+MACD)", trend_curve),
        ("三档+冷却协议", ladder_curve),
        ("切换系统", sw_curve),
        ("组合(20%底仓+80%切换)", combo_curve),
    ]
    perm_curve = None
    if higher_df is not None:
        weekly_perm = higher_permission(df, higher_df, "MACD")
        if weekly_perm is not None:
            perm_curve = switch_backtest(df, hold, rmask & weekly_perm)[
                "equity_curve"
            ].iloc[: n + 1]
            rows.insert(4, ("切换+网格方向过滤", perm_curve))

    print("[绩效汇总] 关键策略完整指标（零费率）")
    print("  策略                     | 总收益 |  年化  | 夏普 | 卡尔玛 | 最大回撤")
    for name, curve in rows:
        m = perf_metrics(curve.reset_index(drop=True), ppy)
        sharpe = f"{m['sharpe']:.2f}" if m["sharpe"] is not None else "  --"
        calmar = f"{m['calmar']:.2f}" if m["calmar"] is not None else "  --"
        print(
            f"  {name:<24} | {fmt_pct(m['total']):>6} | {m['annual']:+6.1%}"
            f" | {sharpe:>4} | {calmar:>6} | {fmt_pct(m['max_dd']):>7}"
        )


def signal_density_report(
    df: pd.DataFrame, higher_df: pd.DataFrame | None, n_days: int,
) -> None:
    """信号密度分析：最近 N 天内多少策略触发了买入/卖出信号。

    多策略密集同向触发（≥5 个）= 市场变盘前兆。扫描全部策略
    （单指标/双确认/三确认/对偶/过滤版/三档/切换）+ 背离预警，
    统计最近 N 天内每个策略的触发事件，输出密集度评级与明细。
    """
    bulls = signal_bulls(df)
    cols = ma_columns(df)
    if len(cols) < 2 or "MACD" not in bulls:
        print("[信号密度] 信号源不足，跳过")
        return

    pair = f"MA{cols[0][2:]}x{cols[-1][2:]}"
    combo = bulls[pair] & bulls["MACD"]
    rmask = regime_mask(df)

    triggers = collect_density_triggers(df, bulls, combo, rmask, higher_df, n_days)

    # 汇总
    buys = [t for t in triggers if t[1] == "买入"]
    sells = [t for t in triggers if t[1] == "卖出"]
    warns = [t for t in triggers if t[1] == "预警"]
    total = len(buys) + len(sells)
    rating = ("高度密集" if total >= 5 else "中等密集" if total >= 3
              else "零星" if total >= 1 else "无信号")

    cutoff = df["date"].iloc[-1] - pd.Timedelta(days=n_days)
    last_date = df["date"].iloc[-1]
    print(f"[信号密度] 最近 {n_days} 天（{cutoff.date()} ~ {last_date.date()}）")
    print(f"  买入 {len(buys)} / 卖出 {len(sells)} / 预警 {len(warns)}"
          f"  →  密集度: {rating}")
    if triggers:
        triggers.sort(key=lambda t: t[2], reverse=True)
        print("  明细（按日期倒序）:")
        for name, direction, date in triggers:
            print(f"    {date.strftime('%Y-%m-%d')}  {direction}  {name}")

    # 信号密度策略回测
    d_hold = density_hold(df, higher_df, n_days=5, threshold=3)
    if d_hold.any():
        r = state_backtest(df, d_hold)
        m = perf_metrics(r["equity_curve"])
        print(f"  [密度策略] 回看5天/阈值3  {r['trades']} 笔  胜率 "
              f"{r['win_rate']:.0%}  收益 {r['total_return']:+.1%}"
              f"  回撤 {r['max_dd']:.1%}  年化 {m['annual']:+.1%}"
              f"  夏普 {m['sharpe']:.2f}  卡尔玛 {m['calmar']:.2f}")
    else:
        print("  [密度策略] 回看5天/阈值3  无触发")


def density_hold(
    df: pd.DataFrame, higher_df: pd.DataFrame | None,
    n_days: int = 5, threshold: int = 3,
) -> pd.Series:
    """信号密度策略持仓掩码。

    逻辑：每个交易日回看最近 n_days 根 K 线（含当日），统计全部掩码型
    策略的买入触发（多头 False→True）与卖出触发（True→False）数量。
    买入触发 >= threshold → 应持仓；卖出触发 >= threshold → 应空仓。
    其余日子维持当前状态不变（类似状态机：信号密集时翻转，否则保持）。

    返回当日判定掩码（调用方需 shift(1) 保证无未来函数）。
    三档/切换等非掩码型策略不纳入（它们的触发与掩码型高度耦合，
    纳入会重复计数）。
    """
    bulls = signal_bulls(df)
    cols = ma_columns(df)
    if len(cols) < 2 or "MACD" not in bulls:
        return pd.Series(False, index=df.index)

    pair = f"MA{cols[0][2:]}x{cols[-1][2:]}"
    combo = bulls[pair] & bulls["MACD"]

    # 复用 build_all_masks 获取全部掩码型策略的持仓掩码
    all_masks = [hold for _, hold in build_all_masks(df, bulls, combo, higher_df)]

    n = len(df)
    buy_counts = pd.Series(0, index=df.index, dtype=int)
    sell_counts = pd.Series(0, index=df.index, dtype=int)
    for mask in all_masks:
        mask = mask.astype(bool)
        prev = mask.shift(1, fill_value=mask.iloc[0])
        buy_counts += ((~prev) & mask).astype(int)
        sell_counts += (prev & (~mask)).astype(int)

    buy_roll = buy_counts.rolling(n_days, min_periods=1).sum()
    sell_roll = sell_counts.rolling(n_days, min_periods=1).sum()

    hold = pd.Series(False, index=df.index)
    for i in range(n):
        if buy_roll.iloc[i] >= threshold:
            hold.iloc[i] = True
        elif sell_roll.iloc[i] >= threshold:
            hold.iloc[i] = False
        elif i > 0:
            hold.iloc[i] = hold.iloc[i - 1]
    return hold


def analyze(symbol: str, period: str, limit: int, density: int = 5) -> None:
    """输出单个周期的完整分析报告。"""
    df = load(symbol, period)
    name = PERIOD_NAMES.get(period, period)
    span = f"{df['date'].iloc[0].date()} ~ {df['date'].iloc[-1].date()}"
    print(f"===== {symbol_label(symbol)} {name} ({span}, {len(df)} 根K线) =====")

    # 周期许可：加载高周期数据（缺失时降级为仅趋势过滤）
    higher_df = None
    higher_period = PERIOD_HIGHER.get(period)
    if higher_period:
        try:
            higher_df = load(symbol, higher_period)
        except FileNotFoundError:
            pass
    regime_report(df)
    print()
    snapshot(df)
    print()
    signal_stats(df)
    print()
    recent_signals(df, limit)
    print()
    signal_density_report(df, higher_df, density)
    print()
    multi_combo_report(df)
    print()
    filter_report(df, higher_df)
    print()
    signal_quality_report(df)
    print()
    ladder_report(df)
    print()
    grid_report(df, higher_df)
    print()
    perf_report(df, higher_df, period)
    print()


def main() -> None:
    parser = argparse.ArgumentParser(description="金叉死叉分析报告")
    parser.add_argument("--symbol", default="562500", help="证券代码")
    parser.add_argument(
        "--period",
        default=None,
        help="daily/weekly/1/5/15/30/60；缺省则分析该标的全部已缓存周期",
    )
    parser.add_argument("--limit", type=int, default=10, help="最近信号显示条数")
    parser.add_argument("--density", type=int, default=5, help="信号密度分析天数（默认 5）")
    args = parser.parse_args()

    if args.period:
        periods = [args.period]
    else:
        periods = [
            p.stem.split("_", 1)[1] for p in DATA_DIR.glob(f"{args.symbol}_*.csv")
        ]
        periods.sort(key=lambda p: PERIOD_ORDER.get(p, 99))
    if not periods:
        raise SystemExit(
            f"data/ 下没有 {args.symbol} 的数据，请先运行 fetch_data.py"
        )

    for period in periods:
        analyze(args.symbol, period, args.limit, args.density)


if __name__ == "__main__":
    main()
