"""随机横截面面板回测：随机抽样 ETF 与个股，跑同一套 5 年策略对比。

目的：三标的（562500 / 000001 / 000905）验证出的结构性结论，放到随机
横截面上统计存活率——看结论是普遍规律，还是单一标的的运气。

数据通道（新浪；东财接口 403 时的替代路径，与 fetch_data.py 的指数
降级路径同源）：

- ETF 清单 / 行情：fund_etf_category_sina / fund_etf_hist_sina（不复权，
  有分红折算的 ETF 在除息日附近会有未复权跳空——随机样本中如出现
  债券/货币类 ETF 需注意此口径）
- 个股清单 / 行情：stock_info_a_code_name / stock_zh_a_daily(qfq 前复权)
- 周线由日线按 W-FRI 重采样（与 fetch_data.fetch_index_raw 同口径）

抽样与窗口：

- 固定随机种子，洗牌后顺序取（不放回抽样），失败/不合格自动换下一个
- 幸存者偏差：从当前上市清单抽样，窗口内退市者缺席
- 上市须早于 2021-04-21（窗口前 ≥ 6 个月指标预热），行数 ≥ 1250
- 拉取 2020-10-21 起日线（多 1 年预热），回测窗口 2021-10-21 ~ 末
  ——与 562500（2021-12 上市）的持有期基本对齐
- 缓存 data/{sh|sz}{code}_{daily|weekly}.csv，与既有缓存同构可复用

回测组合（全部无未来函数、零费率，与 analyze.py 逐项同口径）：

持有 / MA10x20 / MACD / 双确认(MA5x20+MACD) / 三档×3（信号清仓 | 纯移动
止损 | 冷却协议）/ 切换系统 / 切换+网格方向过滤(周线MACD) / 纯网格

末尾输出「结论存活统计」：三标的验证出的每条结构性结论，在随机
横截面上对多少个标的成立。

用法（在本目录下）：

    .venv/bin/python panel.py                    # 10 ETF + 5 个股，seed=20261021
    .venv/bin/python panel.py --etfs 0 --stocks 8
    .venv/bin/python panel.py --seed 42 --etfs 10 --stocks 0
"""

from __future__ import annotations

import argparse
import random
import time

import akshare as ak
import pandas as pd

from analyze import (
    buy_and_hold,
    higher_permission,
    load,
    regime_mask,
    signal_bulls,
    state_backtest,
)
from fetch_data import DATA_DIR, enrich, normalize
from grid import simple_grid_backtest, switch_backtest
from ladder import ladder_backtest

SEED = 20261021
FETCH_START = pd.Timestamp("2020-10-21")    # 行情拉取起点（含 1 年指标预热）
WARMUP_DEADLINE = pd.Timestamp("2021-04-21")  # 上市须早于此
WINDOW_START = pd.Timestamp("2021-10-21")   # 5 年回测窗口起点
MIN_ROWS = 1250                             # 拉取区间内最少行数（约 5.5 年）

STRATEGIES = [
    "持有", "MA10x20", "MACD", "双确认",
    "三档信号", "三档纯移", "三档冷却",
    "切换", "切换滤", "纯网格",
]

# 结论存活检查表：三标的矩阵中验证过的结构性结论
CHECKS = [
    ("双确认收益>持有", lambda r: r["双确认"][0] > r["持有"][0]),
    ("双确认回撤更浅(优于持有)", lambda r: r["双确认"][1] > r["持有"][1]),
    ("纯网格亏损", lambda r: r["纯网格"][0] < 0),
    ("切换跑输纯趋势", lambda r: r["切换"][0] < r["双确认"][0]),
    ("网格方向过滤正贡献", lambda r: "切换滤" in r and r["切换滤"][0] > r["切换"][0]),
    ("三档冷却回撤<单笔", lambda r: r["三档冷却"][1] > r["双确认"][1]),
    ("冷却修复收益(冷却>裸三档)", lambda r: r["三档冷却"][0] > r["三档信号"][0]),
]


def fetch_sina_daily(sinacode: str, is_etf: bool) -> pd.DataFrame:
    """新浪全量历史 -> 截取预热期起点后的 6 列行情。"""
    if is_etf:
        h = ak.fund_etf_hist_sina(symbol=sinacode)
    else:
        h = ak.stock_zh_a_daily(symbol=sinacode, adjust="qfq")
    df = h[["date", "open", "high", "low", "close", "volume"]].copy()
    df["date"] = pd.to_datetime(df["date"])
    df = df[df["date"] >= FETCH_START]
    df = df.dropna(subset=["open", "high", "low", "close"])
    df = df[df["close"] > 0]
    return df.reset_index(drop=True)


def cache_instrument(sinacode: str, daily_raw: pd.DataFrame) -> None:
    """行情 -> enrich 指标/信号列 -> 写 daily/weekly 两个缓存 CSV。"""
    daily = enrich(normalize(daily_raw), (5, 10, 20), (12, 26, 9))
    daily.to_csv(DATA_DIR / f"{sinacode}_daily.csv", index=False)
    weekly_raw = (
        daily_raw.set_index("date")
        .resample("W-FRI")
        .agg({"open": "first", "high": "max", "low": "min",
              "close": "last", "volume": "sum"})
        .dropna(subset=["close"])
        .reset_index()
    )
    weekly = enrich(normalize(weekly_raw), (5, 10, 20), (12, 26, 9))
    weekly.to_csv(DATA_DIR / f"{sinacode}_weekly.csv", index=False)


def etf_universe_sina() -> dict[str, str]:
    """新浪 ETF 清单（代码已带 sh/sz 前缀）。"""
    df = ak.fund_etf_category_sina(symbol="ETF基金")
    return dict(zip(df["代码"].astype(str), df["名称"].astype(str)))


def stock_universe_sina() -> dict[str, str]:
    """A 股清单映射为新浪代码；剔除退市整理/北交所（新浪日接口不支持）。"""
    df = ak.stock_info_a_code_name()
    out: dict[str, str] = {}
    for code, name in zip(df["code"].astype(str), df["name"].astype(str)):
        if "退" in name or name.startswith("PT"):
            continue
        if code.startswith("6"):
            out[f"sh{code}"] = name
        elif code.startswith(("0", "3")):
            out[f"sz{code}"] = name
    return out


def pick_samples(
    universe: dict[str, str], n: int, rng: random.Random, is_etf: bool
) -> list[tuple[str, str]]:
    """洗牌后顺序抽样：拉取失败/上市过晚/行数不足自动换下一个。"""
    codes = sorted(universe)
    rng.shuffle(codes)
    chosen: list[tuple[str, str]] = []
    for code in codes:
        if len(chosen) >= n:
            break
        name = universe[code]
        # 已有缓存直接复用（同种子重跑不联网）
        cached = DATA_DIR / f"{code}_daily.csv"
        if cached.exists():
            dates = pd.read_csv(cached, usecols=["date"])["date"]
            first = pd.to_datetime(dates).iloc[0]
            if first <= WARMUP_DEADLINE and len(dates) >= MIN_ROWS:
                chosen.append((code, name))
                print(f"  [缓存] {code} {name}（首日 {first.date()}）")
                continue
        try:
            raw = fetch_sina_daily(code, is_etf)
        except Exception as exc:  # noqa: BLE001
            print(f"  [弃] {code} {name}: 拉取失败 {type(exc).__name__}")
            continue
        if raw.empty:
            print(f"  [弃] {code} {name}: 无数据")
            continue
        first = pd.Timestamp(raw["date"].iloc[0])
        if first > WARMUP_DEADLINE or len(raw) < MIN_ROWS:
            print(f"  [弃] {code} {name}: 上市过晚/行数不足"
                  f"（首日 {first.date()}，{len(raw)} 行）")
            continue
        cache_instrument(code, raw)
        chosen.append((code, name))
        print(f"  [取] {code} {name}（首日 {first.date()}，{len(raw)} 行）")
        time.sleep(0.3)
    return chosen


def battery(symbol: str) -> dict:
    """单标的全套回测：截取 5 年窗口（指标已在预热期收敛）。"""
    df = load(symbol, "daily")
    df = df[df["date"] >= WINDOW_START].reset_index(drop=True)
    weekly = load(symbol, "weekly")
    bulls = signal_bulls(df)
    hold = bulls["MA5x20"] & bulls["MACD"]

    def sr(r: dict) -> tuple[float, float]:
        return (r["total_return"], r["max_dd"])

    out: dict[str, tuple[float, float]] = {}
    bh, bhdd = buy_and_hold(df)
    out["持有"] = (bh, bhdd)
    out["MA10x20"] = sr(state_backtest(df, bulls["MA10x20"]))
    out["MACD"] = sr(state_backtest(df, bulls["MACD"]))
    out["双确认"] = sr(state_backtest(df, hold))
    out["三档信号"] = sr(ladder_backtest(df, hold, exit_on_signal=True))
    out["三档纯移"] = sr(ladder_backtest(df, hold, exit_on_signal=False))
    out["三档冷却"] = sr(
        ladder_backtest(df, hold, exit_on_signal=True, cooldown=True)
    )
    rmask = regime_mask(df)
    if rmask is None:
        raise ValueError(f"{symbol} 缺 regime 列")
    out["切换"] = sr(switch_backtest(df, hold, rmask))
    wp = higher_permission(df, weekly, "MACD")
    if wp is not None:
        out["切换滤"] = sr(switch_backtest(df, hold, rmask & wp))
    out["纯网格"] = sr(simple_grid_backtest(df))
    out["_meta"] = {
        "rows": len(df),
        "start": str(df["date"].iloc[0].date()),
        "end": str(df["date"].iloc[-1].date()),
    }
    return out


def print_table(results: list[dict], field: int) -> None:
    """field=0 收益表，field=1 回撤表（单位 %）。"""
    title = "总收益" if field == 0 else "最大回撤"
    print(f"\n[{title} %] 窗口 2021-10-21 ~ 2026-10-21（零费率）")
    head = f"  {'代码':<10}{'名称':<14}{'类型':<5}" + "".join(
        f"{s:>9}" for s in STRATEGIES
    )
    print(head)
    for r in results:
        line = f"  {r['代码']:<10}{r['名称'][:11]:<13}{r['类型']:<6}"
        for s in STRATEGIES:
            v = r.get(s)
            line += f"{v[field] * 100:>9.1f}" if v else f"{'--':>9}"
        print(line)
    # 随机样本（不含参考行）中位数
    panel = [r for r in results if r["类型"] != "参考"]
    if panel:
        line = f"  {'(中位数)':<10}{'随机样本':<14}{'':<6}"
        for s in STRATEGIES:
            vals = [r[s][field] for r in panel if r.get(s)]
            line += f"{pd.Series(vals).median() * 100:>9.1f}" if vals else f"{'--':>9}"
        print(line)


def print_survival(results: list[dict]) -> None:
    print("\n[结论存活统计] 三标的矩阵的结构性结论 x 随机横截面")
    panel = [r for r in results if r["类型"] != "参考"]
    print(f"  {'结论':<24} | {'ETF':>6} | {'股票':>6} | {'合计':>6}")
    for label, fn in CHECKS:
        sub = {"ETF": [], "股票": []}
        for r in panel:
            if r["类型"] in sub:
                sub[r["类型"]].append(fn(r))
        e = sum(sub["ETF"])
        s = sum(sub["股票"])
        print(f"  {label:<24} | {e:>3}/{len(sub['ETF']):<2} | {s:>3}/{len(sub['股票']):<2} | {e + s:>3}/{len(panel):<2}")


def main() -> None:
    parser = argparse.ArgumentParser(description="随机横截面面板回测")
    parser.add_argument("--etfs", type=int, default=10, help="随机 ETF 数量")
    parser.add_argument("--stocks", type=int, default=5, help="随机个股数量")
    parser.add_argument("--seed", type=int, default=SEED, help="随机种子")
    args = parser.parse_args()
    rng = random.Random(args.seed)

    print(f"[抽样] seed={args.seed}，目标 {args.etfs} ETF + {args.stocks} 个股"
          f"（幸存者偏差：仅当前上市品种）")
    samples: list[tuple[str, str, str]] = []
    if args.etfs:
        etf_u = etf_universe_sina()
        print(f"[ETF 池] 新浪清单 {len(etf_u)} 只")
        samples += [("ETF", c, n) for c, n in pick_samples(etf_u, args.etfs, rng, True)]
    if args.stocks:
        stock_u = stock_universe_sina()
        print(f"[个股池] 沪深 A 股 {len(stock_u)} 只（剔除退市整理/北交所）")
        samples += [("股票", c, n) for c, n in pick_samples(stock_u, args.stocks, rng, False)]

    if not samples:
        raise SystemExit("未抽到任何样本")

    results: list[dict] = []
    for kind, code, name in samples:
        res = battery(code)
        row = {"代码": code, "名称": name, "类型": kind, **res, **res.pop("_meta")}
        results.append(row)
        print(f"[回测] {code} {name} 完成")

    # 参考锚点：已缓存的三标的（000001/000905 同窗口截取，562500 自上市）
    for code, name in [("000905", "中证500(同窗)"), ("000001", "上证指数(同窗)"),
                       ("562500", "500增强ETF")]:
        try:
            res = battery(code)
            results.append({"代码": code, "名称": name, "类型": "参考",
                            **res, **res.pop("_meta")})
        except FileNotFoundError:
            pass

    print_table(results, 0)
    print_table(results, 1)
    print_survival(results)

    # 结果落盘（长表：每标的两行收益/回撤）
    flat = []
    for r in results:
        rec = {"代码": r["代码"], "名称": r["名称"], "类型": r["类型"],
               "行数": r["rows"], "窗口": f"{r['start']}~{r['end']}"}
        for s in STRATEGIES:
            if r.get(s):
                rec[f"{s}_收益"] = round(r[s][0] * 100, 1)
                rec[f"{s}_回撤"] = round(r[s][1] * 100, 1)
        flat.append(rec)
    out = DATA_DIR / "panel_results.csv"
    pd.DataFrame(flat).to_csv(out, index=False)
    print(f"\n[done] 明细已写入 {out}")


if __name__ == "__main__":
    main()
