"""Pooled 事件研究：信号事件的横截面收益统计（主线 B）。

口径（沿用 GUIDE 铁律，无未来函数）：
    T 日收盘产生信号 → T+1 开盘成交 → 第 N 个交易日收盘平仓
    ret    = close[T+N] / open[T+1] - 1
    excess = ret - 同期池内等权收益      ← 截面去均值，剔除市场共同因子

为什么必须去均值：pooled 后所有事件共享同一段市场历史，若不减掉市场收益，
统计的其实是"这段时间市场涨没涨"，而不是"信号有没有用"。

显著性：按日期聚类的 t 值——同一天的多个事件先平均成 1 个观测（消除截面
相关），再对日期序列做 t 检验。这是 Petersen(2009) 思路的简化版。
注意：本文件只做预设维度的分层，不做任何"挑最优维度"，但仍需自行做多重
检验校正（见 README 的 deflated 说明）。

用法：
    .venv/bin/python event.py
    .venv/bin/python event.py --holds 5,10,20,60 --group regime
"""

from __future__ import annotations

import argparse
import json
import math
from pathlib import Path

import pandas as pd

from analyze import load, regime_components, signal_bulls
from scan import UNIVERSE_PATH, load_ok, load_universe, valid_mask

HOLDS = (5, 10, 20, 60)
SIGNAL_COLS = {
    "ma_5x10_signal": "MA5x10",
    "ma_5x20_signal": "MA5x20",
    "ma_10x20_signal": "MA10x20",
    "macd_signal": "MACD",
}
DUAL_SIGNALS = ("MA5x20", "MACD")


# ── 面板与事件 ──────────────────────────────────────────────
def price_panels(symbols: list[dict], period: str) -> tuple[pd.DataFrame, pd.DataFrame]:
    o: dict[str, pd.Series] = {}
    c: dict[str, pd.Series] = {}
    for s in symbols:
        df = load_ok(s["code"], period)
        if df is None:
            continue
        o[s["code"]] = df["open"]
        c[s["code"]] = df["close"]
    return pd.DataFrame(o).sort_index(), pd.DataFrame(c).sort_index()


def regime_labels(df: pd.DataFrame) -> pd.Series | None:
    """逐日 regime 标签：震荡 / 趋势↑ / 趋势↓ / 过渡。"""
    comp = regime_components(df)
    if comp is None or not {"pdi", "mdi"}.issubset(df.columns):
        return None
    adx = df["adx"]
    trend = (comp["score"] == 0) | (adx >= 35.0)
    up = df["pdi"] > df["mdi"]
    label = pd.Series("过渡", index=df.index)
    label[comp["score"] == 3] = "震荡"
    label[trend & up] = "趋势↑"
    label[trend & ~up] = "趋势↓"
    return label


def collect_events(symbols: list[dict], period: str) -> pd.DataFrame:
    """汇总池内所有信号事件：date / code / source / kind。"""
    rows = []
    for s in symbols:
        df = load_ok(s["code"], period)
        if df is None:
            continue
        vm = valid_mask(df)
        for col, name in SIGNAL_COLS.items():
            if col not in df.columns:
                continue
            for kind in ("golden", "death"):
                hit = (df[col] == kind) & vm
                for d in df.index[hit]:
                    rows.append((d, s["code"], name, kind))
        # 双确认翻转（状态由空转多 / 由多转空）
        bulls = signal_bulls(df)
        if all(k in bulls for k in DUAL_SIGNALS):
            dual = (bulls[DUAL_SIGNALS[0]] & bulls[DUAL_SIGNALS[1]]).astype(float)
            dual = dual.where(vm)
            prev = dual.shift(1)
            for d in df.index[(dual == 1) & (prev == 0)]:
                rows.append((d, s["code"], "双确认", "turn_bull"))
            for d in df.index[(dual == 0) & (prev == 1)]:
                rows.append((d, s["code"], "双确认", "turn_bear"))
    ev = pd.DataFrame(rows, columns=["date", "code", "source", "kind"])
    return ev.sort_values("date").reset_index(drop=True)


def attach_returns(ev: pd.DataFrame, symbols: list[dict], period: str,
                   holds: tuple[int, ...] = HOLDS) -> pd.DataFrame:
    """给每个事件挂上 ret / mkt / excess（各持有期）与当日 regime。"""
    O, C = price_panels(symbols, period)
    ev = ev.set_index(["date", "code"])

    for n in holds:
        entry = O.shift(-1)            # T+1 开盘
        exit_ = C.shift(-n)            # 第 N 个交易日收盘
        ret = (exit_ / entry - 1.0)
        mkt = ret.mean(axis=1)         # 池内等权（同口径，同进出场时点）
        ev[f"ret{n}"] = ret.stack().reindex(ev.index)
        ev[f"mkt{n}"] = mkt.reindex(ev.index.get_level_values("date")).to_numpy()
        ev[f"exc{n}"] = ev[f"ret{n}"] - ev[f"mkt{n}"]

    lab: dict[str, pd.Series] = {}
    for s in symbols:
        df = load_ok(s["code"], period)
        if df is None:
            continue
        r = regime_labels(df)
        if r is not None:
            lab[s["code"]] = r
    if lab:
        panel = pd.DataFrame(lab)
        idx = ev.index
        ev["regime"] = [
            panel.loc[d, c] if (d in panel.index and c in panel.columns
                                and pd.notna(panel.loc[d, c])) else "n/a"
            for d, c in idx
        ]
    return ev.reset_index()


# ── 统计 ────────────────────────────────────────────────────
def cluster_t(sub: pd.DataFrame, col: str) -> float | None:
    """按日期聚类的 t 值：同日事件先平均成 1 个观测，再对日期序列做 t 检验。"""
    daily = sub.groupby("date")[col].mean().dropna()
    if len(daily) < 10 or daily.std(ddof=1) == 0:
        return None
    return daily.mean() / (daily.std(ddof=1) / math.sqrt(len(daily)))


def summarize(ev: pd.DataFrame, holds: tuple[int, ...] = HOLDS,
              group: str | None = None, prefix: str = "exc") -> pd.DataFrame:
    """按 key 汇总。prefix=exc 看选股能力（超额），prefix=ret 看择时能力（原始）。"""
    keys = ["source", "kind"] + ([group] if group else [])
    out = []
    for key, sub in ev.groupby(keys, dropna=False):
        key = key if isinstance(key, tuple) else (key,)
        row = dict(zip(keys, key))
        row["n"] = len(sub)
        row["days"] = sub["date"].nunique()
        for n in holds:
            col = f"{prefix}{n}"
            v = sub[col].dropna()
            if v.empty:
                continue
            row[f"{col}_mean"] = v.mean()
            row[f"{col}_med"] = v.median()
            row[f"win{n}"] = (v > 0).mean()
            row[f"t{n}"] = cluster_t(sub, col)
        out.append(row)
    return pd.DataFrame(out)


def cluster_t_diff(ev: pd.DataFrame, source: str, col: str,
                   pos: str, neg: str) -> tuple[float | None, float | None]:
    """两组事件的差值检验：pos 组 − neg 组，按日期聚类（Welch）。

    为什么必须做这个：单组均值显著为正只能说明"这段市场在涨"（本样本所有
    ret60 都为正、t 高达 3.14，但金叉死叉同时显著）——真正衡量信号有无区分
    度的是两组之差。
    """
    g = ev[(ev["source"] == source) & (ev["kind"] == pos)]
    d = ev[(ev["source"] == source) & (ev["kind"] == neg)]
    dg = g.groupby("date")[col].mean().dropna()
    dd = d.groupby("date")[col].mean().dropna()
    if len(dg) < 10 or len(dd) < 10:
        return None, None
    diff = float(dg.mean() - dd.mean())
    se = math.sqrt(dg.var(ddof=1) / len(dg) + dd.var(ddof=1) / len(dd))
    return diff, (diff / se if se > 0 else None)


KIND_PAIR = {
    "MA5x10": ("golden", "death"),
    "MA5x20": ("golden", "death"),
    "MA10x20": ("golden", "death"),
    "MACD": ("golden", "death"),
    "双确认": ("turn_bull", "turn_bear"),
}


def print_diff(ev: pd.DataFrame, holds: tuple[int, ...], prefix: str = "exc") -> None:
    """差值表：看多信号减看空信号，>0 且 t 大才算信号有区分度。"""
    print(f"\n差值检验  {prefix}(看多) − {prefix}(看空)，按日期聚类 Welch t")
    print(f"{'信号':<8}" + "".join(f"{'diff' + str(n):>10}{'t' + str(n):>8}" for n in holds))
    print("-" * (8 + 18 * len(holds)))
    for src, (pos, neg) in KIND_PAIR.items():
        if src not in set(ev["source"]):
            continue
        line = f"{src:<8}"
        for n in holds:
            diff, t = cluster_t_diff(ev, src, f"{prefix}{n}", pos, neg)
            line += f"{diff:>+10.2%}" if diff is not None else f"{'-':>10}"
            line += f"{t:>8.2f}" if t is not None else f"{'-':>8}"
        print(line)
    print("diff>0 表示看多信号确实跑赢看空信号；|t|>2 粗略显著（未做多重检验校正）")


def print_summary(df: pd.DataFrame, holds: tuple[int, ...], group: str | None,
                  prefix: str = "exc") -> None:
    if df.empty:
        print("无事件")
        return
    keys = ["source", "kind"] + ([group] if group else [])
    order = {"MA5x10": 0, "MA5x20": 1, "MA10x20": 2, "MACD": 3, "双确认": 4}
    df = df.assign(_o=df["source"].map(order)).sort_values(["_o", "kind"] + ([group] if group else []))

    print(f"\n{'信号':<8}{'类型':<11}" + (f"{group:<7}" if group else "")
          + f"{'n':>6}{'天数':>6}" + "".join(f"{prefix + str(n):>11}" for n in holds))
    print("-" * (31 + (7 if group else 0) + 11 * len(holds)))
    for _, r in df.iterrows():
        line = f"{r['source']:<8}{r['kind']:<11}" + (f"{str(r[group]):<7}" if group else "")
        line += f"{int(r['n']):>6}{int(r['days']):>6}"
        for n in holds:
            m = r.get(f"{prefix}{n}_mean")
            line += f"{m:>+11.2%}" if pd.notna(m) else f"{'-':>11}"
        print(line)

    print(f"\n{'信号':<8}{'类型':<11}" + (f"{group:<7}" if group else "")
          + f"{'胜率':>6}" + "".join(f"{'胜率' + str(n):>9}" for n in holds[1:])
          + "".join(f"{'t' + str(n):>8}" for n in holds))
    for _, r in df.iterrows():
        line = f"{r['source']:<8}{r['kind']:<11}" + (f"{str(r[group]):<7}" if group else "")
        w = r.get(f"win{holds[0]}")
        line += f"{w:>6.1%}" if pd.notna(w) else f"{'-':>6}"
        for n in holds[1:]:
            w = r.get(f"win{n}")
            line += f"{w:>9.1%}" if pd.notna(w) else f"{'-':>9}"
        for n in holds:
            t = r.get(f"t{n}")
            line += f"{t:>8.2f}" if pd.notna(t) else f"{'-':>8}"
        print(line)

    if prefix == "exc":
        print("\nexc = 超额收益（已减同期池内等权收益）= 选股能力，剔除了市场涨跌")
    else:
        print("\nret = 原始收益（未减市场）= 择时能力，含市场涨跌的共同贡献")
    print("t   = 按日期聚类的 t 值（|t|>2 为粗略显著，未做多重检验校正）")
    print("golden/turn_bull 期望为正；death/turn_bear 期望为负")


def main() -> None:
    ap = argparse.ArgumentParser(description="Pooled 事件研究")
    ap.add_argument("--universe", default=str(UNIVERSE_PATH))
    ap.add_argument("--holds", default=",".join(str(x) for x in HOLDS))
    ap.add_argument("--group", default="", choices=["", "regime"],
                    help="分层维度（留空=不分层）")
    ap.add_argument("--metric", default="both", choices=["raw", "excess", "both"],
                    help="raw=原始收益(择时能力) / excess=超额收益(选股能力)")
    ap.add_argument("--save", default="", help="事件明细落盘路径（CSV）")
    args = ap.parse_args()

    holds = tuple(int(x) for x in args.holds.split(","))
    group = args.group or None
    cfg = load_universe(Path(args.universe))
    period = cfg.get("period", "daily")
    equity = [s for s in cfg["symbols"] if s["group"] == "equity"]
    print(f"标的池: equity={len(equity)}  period={period}  holds={holds}")

    ev = collect_events(equity, period)
    print(f"事件总数: {len(ev)}（{ev['date'].nunique()} 个交易日）")
    ev = attach_returns(ev, equity, period, holds)
    if args.save:
        ev.to_csv(args.save, index=False, encoding="utf-8-sig")
        print(f"事件明细已落盘 {args.save}")
    prefixes = {"raw": ["ret"], "excess": ["exc"], "both": ["ret", "exc"]}[args.metric]
    for p in prefixes:
        tag = "原始收益 ret（择时能力）" if p == "ret" else "超额收益 exc（选股能力）"
        print(f"\n===== {tag} =====")
        print_summary(summarize(ev, holds, group, p), holds, group, p)
        print_diff(ev, holds, p)


if __name__ == "__main__":
    main()
