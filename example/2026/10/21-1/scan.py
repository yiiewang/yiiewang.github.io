"""方向与信号截面扫描。

定位：只做「方向」与「信号」的横截面分析，不做持仓管理与调仓动作。
两种模式：

    .venv/bin/python scan.py                 当日截面快照（每标的 regime / 各信号源多空）
    .venv/bin/python scan.py --history       回算 breadth / flow / 离散度 序列

--history 输出 data/breadth_history.csv，是主线 A（市场方向）的数据底座，
供后续的领先-滞后检验、信号分层归因（主线 B）直接取用。

复用 analyze.py 的 load / signal_bulls / regime_components，口径保持单一。
"""

from __future__ import annotations

import argparse
import json
from pathlib import Path

import pandas as pd

from analyze import DATA_DIR, load, regime_components, signal_bulls

ROOT = Path(__file__).resolve().parent
UNIVERSE_PATH = ROOT / "universe.json"
BREADTH_PATH = DATA_DIR / "breadth_history.csv"
SCAN_PATH = DATA_DIR / "scan_history.csv"

# GUIDE §4 默认双确认 = MA5x20 且 MACD
DUAL_SIGNALS = ("MA5x20", "MACD")
DUAL_COLS = ("ma_5x20_signal", "macd_signal")

REGIME_SCORE_MIN_NOTE = "3/3=震荡, 0/3 或 ADX>=35=趋势, 其余=过渡"


# ── 池子与数据加载 ──────────────────────────────────────────
def load_universe(path: Path = UNIVERSE_PATH) -> dict:
    if not path.exists():
        raise SystemExit(f"缺少 {path}，请先创建标的池配置")
    return json.loads(path.read_text(encoding="utf-8"))


def valid_mask(df: pd.DataFrame) -> pd.Series:
    """指标预热期掩码：MA20 与 DEA 都已算出才算有效。

    关键正确性点：预热期 ma 列为 NaN，NaN > NaN 得 False，若直接统计会被
    误当成「空头」而系统性低估 breadth。必须先掩掉。
    """
    cols = [c for c in ("ma20", "dea") if c in df.columns]
    if not cols:
        return pd.Series(True, index=df.index)
    return df[cols].notna().all(axis=1)


def load_ok(code: str, period: str) -> pd.DataFrame | None:
    """按 date 索引加载；缺文件返回 None（不中断整批）。"""
    try:
        df = load(code, period)
    except FileNotFoundError:
        return None
    return df.set_index("date").sort_index()


def bull_panel(symbols: list[dict], period: str) -> dict[str, pd.DataFrame]:
    """各信号源的多头状态面板 {信号名: DataFrame(date × code)}，1.0=多 0.0=空，预热期 NaN。"""
    panels: dict[str, dict[str, pd.Series]] = {}
    for s in symbols:
        df = load_ok(s["code"], period)
        if df is None:
            print(f"  [跳过] {s['code']} 缺少 {period} 数据")
            continue
        vm = valid_mask(df)
        for name, bull in signal_bulls(df).items():
            panels.setdefault(name, {})[s["code"]] = bull.astype(float).where(vm)
    return {k: pd.DataFrame(v).sort_index() for k, v in panels.items()}


def _and(a: pd.DataFrame, b: pd.DataFrame) -> pd.DataFrame:
    """两个状态面板的逻辑与，任一侧 NaN 则结果 NaN。"""
    ok = a.notna() & b.notna()
    return ((a > 0) & (b > 0)).where(ok)


def event_counts(symbols: list[dict], period: str,
                 cols: tuple[str, ...] = DUAL_COLS) -> pd.DataFrame:
    """信号事件计数面板 date × code：golden=+1 / death=-1 的逐日累计（双确认两个信号源合计）。"""
    data: dict[str, pd.Series] = {}
    for s in symbols:
        df = load_ok(s["code"], period)
        if df is None:
            continue
        ev = pd.Series(0.0, index=df.index)
        for c in cols:
            if c not in df.columns:
                continue
            ev = ev + df[c].map({"golden": 1.0, "death": -1.0}).fillna(0.0)
        data[s["code"]] = ev.where(valid_mask(df))
    return pd.DataFrame(data).sort_index()


def ret_panel(symbols: list[dict], period: str, n: int) -> pd.DataFrame:
    """N 日收益面板 date × code（用于横截面离散度）。"""
    data: dict[str, pd.Series] = {}
    for s in symbols:
        df = load_ok(s["code"], period)
        if df is None:
            continue
        data[s["code"]] = df["close"].pct_change(n)
    return pd.DataFrame(data).sort_index()


# ── 模式一：当日截面快照 ────────────────────────────────────
def _state_run(bull: pd.Series) -> int:
    """当前状态已持续多少根 K 线（从最后一段同状态起算）。"""
    bb = bull.dropna()
    if bb.empty:
        return 0
    cur = bb.iloc[-1]
    ne = (bb != cur).values.nonzero()[0]
    start = int(ne[-1]) + 1 if len(ne) else 0
    return len(bb) - start


def _regime_label(df: pd.DataFrame) -> tuple[str, int]:
    comp = regime_components(df)
    last = df.iloc[-1]
    if comp is None:
        return "n/a", -1
    score = int(comp["score"].iloc[-1])
    if score == 3:
        return "震荡", score
    if score == 0 or float(last.get("adx", 0)) >= 35.0:
        up = float(last.get("pdi", 0)) > float(last.get("mdi", 0))
        return f"趋势{'↑' if up else '↓'}", score
    return "过渡", score


def run_snapshot(cfg: dict, write: bool = True) -> pd.DataFrame:
    period = cfg.get("period", "daily")
    symbols = cfg["symbols"]
    rows = []
    for s in symbols:
        df = load_ok(s["code"], period)
        if df is None:
            print(f"  [跳过] {s['code']} 缺少 {period} 数据")
            continue
        last = df.iloc[-1]
        regime, score = _regime_label(df.reset_index())
        row = {
            "date": last.name.date().isoformat(),
            "code": s["code"],
            "name": s.get("name", ""),
            "group": s["group"],
            "close": round(float(last["close"]), 4),
            "regime": regime,
            "score": score,
            "adx": round(float(last["adx"]), 1) if "adx" in df.columns else None,
            "dd": round(float(last["dd"]), 3) if "dd" in df.columns else None,
        }
        bulls = signal_bulls(df)
        for name, b in bulls.items():
            row[f"s_{name}"] = int(bool(b.dropna().iloc[-1])) if not b.dropna().empty else None
            row[f"run_{name}"] = _state_run(b)
        if DUAL_SIGNALS[0] in bulls and DUAL_SIGNALS[1] in bulls:
            dual = _and(bulls[DUAL_SIGNALS[0]].astype(float),
                        bulls[DUAL_SIGNALS[1]].astype(float))
            row["s_dual"] = int(bool(dual.dropna().iloc[-1])) if not dual.dropna().empty else None
        if {"bear_div", "bull_div"}.issubset(df.columns):
            tail = df.tail(10)
            row["div_warn"] = ("顶背离" if tail["bear_div"].any() else
                               "底背离" if tail["bull_div"].any() else "")
        rows.append(row)

    out = pd.DataFrame(rows)
    print_snapshot(out)
    if write and not out.empty:
        upsert(SCAN_PATH, out, keys=["date", "code"])
        print(f"\n已落盘 {SCAN_PATH}（按 date+code 幂等覆盖）")
    return out


def print_snapshot(df: pd.DataFrame) -> None:
    names = [c[2:] for c in df.columns if c.startswith("s_")]
    head = f"{'代码':<10}{'名称':<18}{'组':<8}{'日期':<12}{'收盘':>9}  {'regime':<7}{'分':<4}"
    head += "".join(f"{n:>8}" for n in names)
    print(head)
    print("-" * 80)
    for _, r in df.iterrows():
        line = (f"{r['code']:<10}{str(r['name'])[:16]:<18}{r['group']:<8}"
                f"{r['date']:<12}{r['close']:>9.3f}  {r['regime']:<7}{r['score']:<4}")
        for n in names:
            v = r.get(f"s_{n}")
            if v is None or pd.isna(v):
                line += f"{'-':>8}"
            else:
                run = r.get(f"run_{n}")
                line += f"{'多' if v else '空'}{'' if run is None else f'{int(run)}根':>6}"
        print(line)
    print(f"\nregime 判定: {REGIME_SCORE_MIN_NOTE}")


# ── 模式二：历史 breadth 回算 ───────────────────────────────
def run_history(cfg: dict, out_path: Path = BREADTH_PATH) -> pd.DataFrame:
    period = cfg.get("period", "daily")
    min_n = int(cfg.get("min_n", 10))
    equity = [s for s in cfg["symbols"] if s["group"] == "equity"]
    bench = [s for s in cfg["symbols"] if s["group"] == "bench"]

    print(f"回算标的: equity={len(equity)} bench={len(bench)} period={period}")
    bulls = bull_panel(equity, period)
    if not bulls:
        raise SystemExit("池内没有可加载的标的")

    dates = None
    for p in bulls.values():
        dates = p.index if dates is None else dates.union(p.index)
    dates = dates.sort_values()

    cols: dict[str, pd.Series] = {}
    n_series = None
    for name, p in bulls.items():
        p = p.reindex(dates)
        cols[f"breadth_{name}"] = p.mean(axis=1)
        if n_series is None:
            n_series = p.notna().sum(axis=1)
    if DUAL_SIGNALS[0] in bulls and DUAL_SIGNALS[1] in bulls:
        dual = _and(bulls[DUAL_SIGNALS[0]].reindex(dates),
                    bulls[DUAL_SIGNALS[1]].reindex(dates))
        cols["breadth_dual"] = dual.mean(axis=1)

    ev = event_counts(equity, period).reindex(dates)
    cols["golden"] = (ev > 0).sum(axis=1)
    cols["death"] = (ev < 0).sum(axis=1)
    cols["flow"] = ev.sum(axis=1, min_count=1)

    ret20 = ret_panel(equity, period, 20).reindex(dates)
    cols["disp20"] = ret20.std(axis=1)
    for b in bench:
        df = load_ok(b["code"], period)
        if df is None:
            continue
        cols[f"ret20_{b['code']}"] = df["close"].pct_change(20).reindex(dates)

    out = pd.DataFrame(cols)
    out.insert(0, "n", n_series)

    # 样本数不足的交易日：breadth 类置空，避免上市进度不同造成曲线漂移
    breadth_cols = [c for c in out.columns if c.startswith("breadth_")]
    out.loc[out["n"] < min_n, breadth_cols] = pd.NA

    out_path.parent.mkdir(parents=True, exist_ok=True)
    out.to_csv(out_path, float_format="%.6f", encoding="utf-8-sig")
    print_history_summary(out, min_n)
    print(f"\n已落盘 {out_path}（{len(out)} 行）")
    return out


def print_history_summary(out: pd.DataFrame, min_n: int) -> None:
    valid = out[out["n"] >= min_n]
    print(f"\n有效交易日 {len(valid)} / {len(out)}（min_n={min_n}）"
          f"  区间 {valid.index.min().date()} ~ {valid.index.max().date()}")

    bd = valid["breadth_dual"].dropna() if "breadth_dual" in valid else pd.Series(dtype=float)
    if not bd.empty:
        print(f"\nbreadth_dual（双确认多头占比）")
        print(f"  最近有效 {bd.iloc[-1]:.1%} @ {bd.index[-1].date()}（n={int(valid.loc[bd.index[-1], 'n'])}）"
              f" | 均值 {bd.mean():.1%} | 中位 {bd.median():.1%}")
        print(f"  分位 10%={bd.quantile(.1):.1%} 25%={bd.quantile(.25):.1%} "
              f"75%={bd.quantile(.75):.1%} 90%={bd.quantile(.9):.1%}")
        print(f"  最高 {bd.max():.1%} @ {bd.idxmax().date()} | "
              f"最低 {bd.min():.1%} @ {bd.idxmin().date()}")
        hi = (bd >= 0.8).sum()
        lo = (bd <= 0.2).sum()
        print(f"  >=80% 的天数 {hi}（{hi / len(bd):.1%}）| <=20% 的天数 {lo}（{lo / len(bd):.1%}）")

    # 领先-滞后：breadth 与基准 20 日收益的相关（breadth 领先 k 日）
    bench_cols = [c for c in valid.columns if c.startswith("ret20_")]
    if not bd.empty and bench_cols:
        print("\n相关系数 corr(breadth_dual, 基准20日收益) —— breadth 领先 k 个交易日")
        print(f"  {'k':>4}" + "".join(f"{c:>16}" for c in bench_cols))
        for k in (0, 5, 10, 20):
            row = f"  {k:>4}"
            for c in bench_cols:
                pair = pd.concat([bd.shift(k), valid[c]], axis=1).dropna()
                row += f"{pair.corr().iloc[0, 1]:>16.3f}" if len(pair) > 30 else f"{'-':>16}"
            print(row)


# ── 落盘（幂等） ────────────────────────────────────────────
def upsert(path: Path, new: pd.DataFrame, keys: list[str]) -> None:
    if path.exists():
        old = pd.read_csv(path, dtype=str)
        merged = pd.concat([old, new.astype(str)], ignore_index=True)
        merged = merged.drop_duplicates(subset=keys, keep="last")
    else:
        merged = new
    merged.to_csv(path, index=False, encoding="utf-8-sig")


def main() -> None:
    ap = argparse.ArgumentParser(description="方向与信号截面扫描")
    ap.add_argument("--history", action="store_true", help="回算 breadth/flow/离散度 序列")
    ap.add_argument("--universe", default=str(UNIVERSE_PATH), help="标的池配置路径")
    ap.add_argument("--no-write", action="store_true", help="只打印不落盘")
    args = ap.parse_args()

    cfg = load_universe(Path(args.universe))
    if args.history:
        run_history(cfg)
    else:
        run_snapshot(cfg, write=not args.no_write)


if __name__ == "__main__":
    main()
