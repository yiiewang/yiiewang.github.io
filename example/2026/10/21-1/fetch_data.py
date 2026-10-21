"""从 akshare 拉取 A 股/ETF 行情并缓存为本地 CSV。

数据源为东方财富（akshare 封装），支持：

- 日线 / 周线：period = daily / weekly（可指定起止日期，前复权）
- 分钟线    ：period = 1 / 5 / 15 / 30 / 60（东财仅保留近几个交易日，
              1 分钟线约 5 个交易日，接口不接受起止日期）

用法（在本目录下）：

    python3 fetch_data.py --symbol 562500 --period daily
    python3 fetch_data.py --symbol 562500 --period weekly
    python3 fetch_data.py --symbol 562500 --period 1 --macd 5,10,9 --force

缓存为 data/{symbol}_{period}.csv（按时间升序）：前 6 列为行情
（date,open,close,high,low,volume），其余列由 indicators.py 计算追加——
指标（ma5,ma10,ma20,dif,dea,hist,adx,pdi,mdi,cohesion,don_upper,don_lower,
vol_ratio,bias10,dd）、信号（ma_5x10_signal,ma_5x20_signal,ma_10x20_signal,
macd_signal），打开 CSV 即可直接看到金叉/死叉发生在哪一行。已存在则跳过，
--force 强制重拉。之后 analyze.py 离线读取本地缓存，不再联网。

指标参数由命令行注入：--ma 5,20 与 --macd 12,26,9（默认值），indicators.py
保持标准默认值不动。换参数重拉会覆盖同周期缓存。
"""

from __future__ import annotations

import argparse
import datetime as dt
from pathlib import Path

import akshare as ak
import pandas as pd

from indicators import (
    adx,
    bias,
    donchian,
    drawdown,
    ma,
    ma_cohesion,
    macd,
    macd_signals,
    ma_signals,
    volume_ratio,
)

# 场内 ETF 代码前缀（沪市 51/56/58，深市 15/16），其余按 A 股股票处理
ETF_PREFIXES = ("15", "16", "51", "56", "58")

VALID_PERIODS = ("daily", "weekly", "1", "5", "15", "30", "60")

# 东财接口的中文列名 -> 统一英文列名
COLUMN_MAP = {
    "日期": "date",
    "时间": "date",
    "开盘": "open",
    "收盘": "close",
    "最高": "high",
    "最低": "low",
    "成交量": "volume",
}

DATA_DIR = Path(__file__).resolve().parent / "data"


def is_etf(symbol: str) -> bool:
    """按代码前缀区分场内 ETF 与 A 股股票。"""
    return symbol.startswith(ETF_PREFIXES)


def parse_ints(arg: str, name: str, min_count: int) -> tuple[int, ...]:
    """解析逗号分隔的整数参数，如 '5,10,9' -> (5, 10, 9)，至少 min_count 个。"""
    try:
        values = tuple(int(x) for x in arg.split(","))
    except ValueError:
        raise SystemExit(
            f"--{name} 格式错误：{arg!r}，应为逗号分隔的整数"
        )
    if len(values) < min_count:
        raise SystemExit(
            f"--{name} 至少需要 {min_count} 个参数，收到 {len(values)} 个：{arg!r}"
        )
    return values


def fetch_raw(
    symbol: str, period: str, start: str, end: str, adjust: str = "qfq"
) -> pd.DataFrame:
    """按周期与品种调用对应的 akshare 接口。"""
    if period in ("daily", "weekly"):
        if is_etf(symbol):
            return ak.fund_etf_hist_em(
                symbol=symbol,
                period=period,
                start_date=start,
                end_date=end,
                adjust=adjust,
            )
        return ak.stock_zh_a_hist(
            symbol=symbol,
            period=period,
            start_date=start,
            end_date=end,
            adjust=adjust,
        )
    # 分钟线：东财接口不接受起止日期，返回其保留的全部历史
    if is_etf(symbol):
        return ak.fund_etf_hist_min_em(symbol=symbol, period=period, adjust=adjust)
    return ak.stock_zh_a_hist_min_em(symbol=symbol, period=period, adjust=adjust)


def normalize(df: pd.DataFrame) -> pd.DataFrame:
    """中文列名转英文，按时间升序，只保留统一列。"""
    out = df.rename(columns=COLUMN_MAP)
    for required in ("date", "close"):
        if required not in out.columns:
            raise ValueError(f"接口返回缺少必需列 {required}：{list(df.columns)}")
    keep = [c for c in ("date", "open", "close", "high", "low", "volume") if c in out.columns]
    out = out[keep].copy()
    if "open" in out.columns:
        # 东财 ETF 分钟线部分行 open 为 0（接口缺陷），用当日 close 修复，
        # 否则回测"次日开盘成交"会得到 0 成本买入的虚假收益
        out["open"] = out["open"].where(out["open"] > 0, out["close"])
    out["date"] = pd.to_datetime(out["date"])
    return out.sort_values("date").reset_index(drop=True)


def enrich(
    df: pd.DataFrame,
    ma_windows: tuple[int, ...],
    macd_params: tuple[int, int, int],
) -> pd.DataFrame:
    """调用 indicators 在行情上追加指标、信号与市场状态列。

    参数由命令行注入（--ma / --macd），indicators 只提供计算能力：
    - ma_windows = 均线窗口序列（至少 2 个，默认 5,10,20），每个窗口
      生成数值列 ma{w}；每对组合生成一列信号 ma_{a}x{b}_signal
      （默认三对：5x10 短期、5x20 中期、10x20 长期）
    - macd_params = (fast, slow, signal)，列名固定 dif/dea/hist

    信号列（交叉行取值 golden/death，其余为空）：
    - ma_{a}x{b}_signal（所有均线两两组合）/ macd_signal

    市场状态（regime）列，参数取各指标默认值（14/20/5/10）：
    - adx / pdi / mdi      趋向强度与方向
    - cohesion             多均线粘合度
    - don_upper / don_lower Donchian 通道（前 20 日高低点，不含当日）
    - vol_ratio            量比（对前 5 日均量）
    - bias10               10 日乖离率
    - dd                   回撤序列
    """
    fast, slow, signal = macd_params
    close = df["close"]
    for w in ma_windows:
        df[f"ma{w}"] = ma(close, w)
    dif, dea, hist = macd(close, fast, slow, signal)
    df["dif"] = dif
    df["dea"] = dea
    df["hist"] = hist
    # 每对均线组合一列信号：默认 5x10(短期) / 5x20(中期) / 10x20(长期)
    windows = sorted(ma_windows)
    for i in range(len(windows)):
        for j in range(i + 1, len(windows)):
            a, b = windows[i], windows[j]
            df[f"ma_{a}x{b}_signal"] = ma_signals(close, a, b)
    df["macd_signal"] = macd_signals(close, fast, slow, signal)

    df["pdi"], df["mdi"], df["adx"] = adx(df["high"], df["low"], close, 14)
    df["cohesion"] = ma_cohesion(close)
    df["don_upper"], df["don_lower"] = donchian(df["high"], df["low"], 20)
    df["vol_ratio"] = volume_ratio(df["volume"], 5)
    df["bias10"] = bias(close, 10)
    df["dd"] = drawdown(close)
    return df


def fetch_and_cache(
    symbol: str,
    period: str,
    start: str,
    end: str,
    force: bool = False,
    ma_windows: tuple[int, ...] = (5, 10, 20),
    macd_params: tuple[int, int, int] = (12, 26, 9),
) -> Path:
    """联网拉取行情并写入本地 CSV 缓存，返回缓存路径。"""
    target = DATA_DIR / f"{symbol}_{period}.csv"
    if target.exists() and not force:
        print(f"[skip] 已存在 {target}（如需重拉加 --force）")
        return target

    print(
        f"[fetch] {symbol} period={period}"
        f" MA={ma_windows} MACD={macd_params} ..."
    )
    raw = fetch_raw(symbol, period, start, end)
    if raw is None or raw.empty:
        raise RuntimeError("接口返回空数据，请检查证券代码与网络")

    df = enrich(normalize(raw), ma_windows, macd_params)
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    df.to_csv(target, index=False)
    print(
        f"[done] {len(df)} 行 -> {target}"
        f"（{df['date'].iloc[0]} ~ {df['date'].iloc[-1]}）"
    )
    return target


def recompute_cache(
    symbol: str,
    period: str,
    ma_windows: tuple[int, ...],
    macd_params: tuple[int, int, int],
) -> Path:
    """从本地缓存重算指标列，不联网。

    两个适用场景：
    - 换指标参数（--ma / --macd）后只需重算，行情没变不必重新请求
    - 数据源限流/断网时的降级路径（前提：本地至少拉取过一次）
    """
    target = DATA_DIR / f"{symbol}_{period}.csv"
    if not target.exists():
        raise FileNotFoundError(
            f"缺少本地缓存 {target}，请先联网拉取一次（去掉 --recompute）"
        )
    raw = pd.read_csv(
        target, usecols=["date", "open", "close", "high", "low", "volume"]
    )
    df = enrich(normalize(raw), ma_windows, macd_params)
    df.to_csv(target, index=False)
    print(
        f"[recompute] {len(df)} 行 -> {target}"
        f"（MA={ma_windows} MACD={macd_params}，未联网）"
    )
    return target


def main() -> None:
    parser = argparse.ArgumentParser(
        description="拉取 A 股/ETF 行情并缓存为本地 CSV"
    )
    parser.add_argument("--symbol", default="562500", help="证券代码，默认 562500")
    parser.add_argument(
        "--period", default="daily", choices=VALID_PERIODS, help="K 线周期"
    )
    parser.add_argument(
        "--start", default="20200101", help="起始日期 YYYYMMDD（仅日线/周线生效）"
    )
    parser.add_argument(
        "--end",
        default=dt.date.today().strftime("%Y%m%d"),
        help="结束日期 YYYYMMDD（仅日线/周线生效）",
    )
    parser.add_argument(
        "--ma",
        default="5,10,20",
        help="均线参数：逗号分隔的窗口（默认 5,10,20，至少 2 个），"
             "每个窗口生成一列，金叉死叉信号取最短/最长两条",
    )
    parser.add_argument(
        "--macd",
        default="12,26,9",
        help="MACD 参数：fast,slow,signal（默认 12,26,9）",
    )
    parser.add_argument(
        "--force", action="store_true", help="忽略本地缓存强制重拉"
    )
    parser.add_argument(
        "--recompute",
        action="store_true",
        help="不联网，从本地缓存的行情列重算指标列（换参数后无需重拉）",
    )
    args = parser.parse_args()

    ma_windows = parse_ints(args.ma, "ma", 2)
    macd_params = parse_ints(args.macd, "macd", 3)
    if len(macd_params) != 3:
        raise SystemExit(
            f"--macd 需要 3 个参数，收到 {len(macd_params)} 个：{args.macd!r}"
        )

    if args.recompute:
        recompute_cache(args.symbol, args.period, ma_windows, macd_params)
        return
    fetch_and_cache(
        args.symbol,
        args.period,
        args.start,
        args.end,
        args.force,
        ma_windows,
        macd_params,
    )


if __name__ == "__main__":
    main()
