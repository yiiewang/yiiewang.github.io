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
vol_ratio,bias10,dd,dif_pos）、信号（ma_5x10_signal,ma_5x20_signal,
ma_10x20_signal,macd_signal）、背离事件掩码（bear_div,bull_div），打开 CSV
即可直接看到金叉/死叉/背离发生在哪一行。已存在则跳过，--force 强制重拉。
之后 analyze.py 离线读取本地缓存，不再联网。

指标参数由命令行注入：--ma 5,20 与 --macd 12,26,9（默认值），indicators.py
保持标准默认值不动。换参数重拉会覆盖同周期缓存。
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import re
import time
from pathlib import Path

import akshare as ak
import pandas as pd
import requests

from indicators import (
    adx,
    bias,
    donchian,
    drawdown,
    ma,
    ma_cohesion,
    macd,
    macd_divergence,
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


def to_weekly(df: pd.DataFrame) -> pd.DataFrame:
    """日线按 W-FRI 重采样为周线（open=首日、high/low=极值、close=末日、volume=求和）。"""
    return (
        df.set_index("date")
        .resample("W-FRI")
        .agg({"open": "first", "high": "max", "low": "min", "close": "last", "volume": "sum"})
        .dropna(subset=["close"])
        .reset_index()
    )


def retry_call(fn, attempts: int = 3, base_delay: float = 2.0):
    """行情接口退避重试：东财/新浪偶发限流（返回空体或 HTML 反爬页），
    akshare 内部 r.json() 抛 JSONDecodeError——多为瞬时故障，指数退避
    重试通常可恢复；重试耗尽抛最后一次异常，由调用方决定降级或报错。"""
    for i in range(attempts):
        try:
            return fn()
        except Exception:  # noqa: BLE001
            if i == attempts - 1:
                raise
            delay = base_delay * 2 ** i
            print(f"[retry] 接口失败（第 {i + 1}/{attempts} 次），{delay:g}s 后重试")
            time.sleep(delay)


def bare_code(symbol: str) -> str:
    """剥离 sh/sz 市场前缀，返回纯数字代码。

    universe.json 与 panel.py 抽样缓存使用带前缀写法（sz159972 / sh511360），
    而各行情接口要求纯数字代码——前缀若不剥离会导致 ETF 被误判成股票。
    """
    return symbol[2:] if symbol[:2].lower() in ("sh", "sz") else symbol


def is_etf(symbol: str) -> bool:
    """按代码前缀区分场内 ETF 与 A 股股票（兼容带 sh/sz 前缀的写法）。"""
    return bare_code(symbol).startswith(ETF_PREFIXES)


# 常见指数名称（指数代码与股票代码存在冲突，如 000001 既是上证指数
# 又是平安银行，指数须经 --index 显式指定，故单独维护，不从 A 股表取）
INDEX_NAMES = {
    "000001": "上证指数", "399001": "深证成指", "399006": "创业板指",
    "000300": "沪深300", "000905": "中证500", "000852": "中证1000",
    "399905": "中证500", "399903": "创业板50", "000016": "上证50",
    "000688": "科创50", "399295": "国证2000",
}

META_PATH = DATA_DIR / "meta.json"


def load_meta() -> dict[str, str]:
    """读取本地证券名称缓存 {代码: 名称}（不存在或损坏返回空字典）。"""
    if not META_PATH.exists():
        return {}
    try:
        return json.loads(META_PATH.read_text(encoding="utf-8"))
    except (ValueError, OSError):
        return {}


def _meta_key(symbol: str, index: bool = False) -> str:
    """缓存键：指数加 idx: 前缀——指数代码与股票冲突（000001 既是
    上证指数又是平安银行），必须分开存，否则互相污染。"""
    return f"idx:{symbol}" if index else symbol


def save_name(symbol: str, name: str, index: bool = False) -> None:
    """写入证券名称到本地缓存（merge 已有内容）。"""
    meta = load_meta()
    meta[_meta_key(symbol, index)] = name
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    META_PATH.write_text(
        json.dumps(meta, ensure_ascii=False, indent=2), encoding="utf-8"
    )


def resolve_name(symbol: str, index: bool = False) -> str | None:
    """解析证券名称：本地缓存 -> 联网查询 -> None（失败静默降级）。

    - 指数：优先内置映射（代码与股票冲突，不能查 A 股表）
    - 股票：新浪 A 股代码表（稳定，5577 只）
    - ETF：东财 ETF 现货表（常被反爬 403，失败返回 None）
    查到即写入 data/meta.json 供后续离线使用（chart.py 只读缓存不联网）。
    """
    meta = load_meta()
    key = _meta_key(symbol, index)
    if key in meta:
        return meta[key]

    name: str | None = None
    try:
        if index:
            name = INDEX_NAMES.get(symbol)
            if name is None:
                df = retry_call(lambda: ak.stock_zh_index_spot_sina())
                hit = df[df["代码"].astype(str).str.contains(symbol)]
                if not hit.empty:
                    name = str(hit.iloc[0]["名称"])
        elif is_etf(symbol):
            df = retry_call(lambda: ak.fund_etf_spot_em())
            hit = df[df["代码"].astype(str) == symbol]
            if not hit.empty:
                name = str(hit.iloc[0]["名称"])
        else:
            df = retry_call(lambda: ak.stock_info_a_code_name())
            hit = df[df["code"].astype(str) == symbol]
            if not hit.empty:
                name = str(hit.iloc[0]["name"]).replace(" ", "")
    except Exception:  # noqa: BLE001
        # 名称仅用于展示，接口失败不阻塞主流程（报告退化为只显示代码）
        name = None

    if name:
        save_name(symbol, name, index=index)
    return name


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


def fetch_index_raw(
    symbol: str, period: str, start: str, end: str
) -> pd.DataFrame | None:
    """
    新浪指数日线接口（东财指数接口反爬 403 时的替代）。

    上证指数代码 000001 与深市股票平安银行冲突，指数必须经 --index
    显式指定，内部映射为 sh{symbol}。周线由日线按 W-FRI 重采样
    （open=首日、high/low=极值、close=末日、volume=求和）。
    """
    try:
        raw = retry_call(lambda: ak.stock_zh_index_daily(symbol=f"sh{symbol}"))
    except Exception as exc:  # noqa: BLE001
        print(f"[error] 指数接口失败: {exc}")
        return None
    df = raw.copy()
    df["date"] = pd.to_datetime(df["date"])
    df = df[
        (df["date"] >= pd.to_datetime(start)) & (df["date"] <= pd.to_datetime(end))
    ]
    if df.empty:
        return None
    if period == "weekly":
        df = to_weekly(df)
    return df


def fetch_etf_sina_raw(
    symbol: str, period: str, start: str, end: str
) -> pd.DataFrame | None:
    """
    新浪 ETF 日线接口（东财接口 403 时的降级路径）。

    - 新浪仅有不复权日线，周线由日线按 W-FRI 重采样
      （open=首日、high/low=极值、close=末日、volume=求和）
    - 沪市 ETF 映射 sh{code}，深市映射 sz{code}
    """
    code = bare_code(symbol)
    prefix = "sz" if code.startswith(("15", "16")) else "sh"
    try:
        raw = retry_call(lambda: ak.fund_etf_hist_sina(symbol=f"{prefix}{code}"))
    except Exception as exc:  # noqa: BLE001
        print(f"[error] 新浪 ETF 接口失败: {exc}")
        return None
    df = raw.copy()
    df["date"] = pd.to_datetime(df["date"])
    df = df[
        (df["date"] >= pd.to_datetime(start)) & (df["date"] <= pd.to_datetime(end))
    ]
    if df.empty:
        return None
    if period == "weekly":
        df = to_weekly(df)
    return df


def fetch_stock_sina_raw(
    symbol: str, period: str, start: str, end: str, adjust: str = "qfq",
) -> pd.DataFrame | None:
    """
    新浪股票日线接口（东财接口反爬时的降级路径）。

    与 ETF 新浪降级不同，新浪股票接口支持 adjust 参数，可保持前复权，
    不改变均线/MACD 在除息附近的口径。周线由日线按 W-FRI 重采样
    （open=首日、high/low=极值、close=末日、volume=求和）。
    沪市映射 sh{code}，深市映射 sz{code}。
    """
    code = bare_code(symbol)
    prefix = "sz" if code.startswith(("0", "3")) else "sh"
    try:
        raw = retry_call(lambda: ak.stock_zh_a_daily(
            symbol=f"{prefix}{code}",
            start_date=start.replace("-", ""),
            end_date=end.replace("-", ""),
            adjust=adjust,
        ))
    except Exception as exc:  # noqa: BLE001
        print(f"[error] 新浪股票接口失败: {exc}")
        return None
    df = raw.copy()
    # 新浪接口列名：date/open/high/low/close/volume（已英文，无需 rename）
    df["date"] = pd.to_datetime(df["date"])
    df = df[
        (df["date"] >= pd.to_datetime(start)) & (df["date"] <= pd.to_datetime(end))
    ]
    if df.empty:
        return None
    if period == "weekly":
        df = to_weekly(df)
    return df


def fetch_raw(
    symbol: str, period: str, start: str, end: str, adjust: str = "qfq",
    index: bool = False,
) -> pd.DataFrame | None:
    """按周期与品种调用对应的 akshare 接口。"""
    if index:
        if period not in ("daily", "weekly"):
            raise SystemExit("指数接口仅支持日线/周线（分钟线不支持）")
        return fetch_index_raw(symbol, period, start, end)
    if period in ("daily", "weekly"):
        if is_etf(symbol):
            try:
                # 限流返回空体（akshare 内 JSON 解析抛错）多为瞬时，先退避重试
                return retry_call(lambda: ak.fund_etf_hist_em(
                    symbol=bare_code(symbol),
                    period=period,
                    start_date=start,
                    end_date=end,
                    adjust=adjust,
                ))
            except Exception as exc:  # noqa: BLE001
                # 重试耗尽仍失败（如数据中心 IP 被持续反爬），降级新浪接口
                # 继续取数（注意：新浪仅不复权数据，均线/MACD 在除息附近
                # 的信号会与前复权不同）
                print(f"[warn] 东财 ETF 接口重试后仍失败({exc})，降级新浪接口（不复权）")
                return fetch_etf_sina_raw(symbol, period, start, end)
        try:
            return retry_call(lambda: ak.stock_zh_a_hist(
                symbol=bare_code(symbol),
                period=period,
                start_date=start,
                end_date=end,
                adjust=adjust,
            ))
        except Exception as exc:  # noqa: BLE001
            # 东财反爬（返回空体/HTML）重试耗尽后降级新浪股票接口
            # （新浪支持前复权，口径与东财一致）
            print(f"[warn] 东财股票接口重试后仍失败({exc})，降级新浪接口")
            return fetch_stock_sina_raw(symbol, period, start, end, adjust)
    # 分钟线：东财接口不接受起止日期，返回其保留的全部历史
    if is_etf(symbol):
        return retry_call(lambda: ak.fund_etf_hist_min_em(
            symbol=bare_code(symbol), period=period, adjust=adjust))
    return retry_call(lambda: ak.stock_zh_a_hist_min_em(
        symbol=bare_code(symbol), period=period, adjust=adjust))


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

    # 背离（预警事件掩码）与零轴位置（DIF>0 的市场状态）
    bear, bull = macd_divergence(close, dif, 20)
    df["bear_div"] = bear
    df["bull_div"] = bull
    df["dif_pos"] = dif > 0
    return df


def sina_quote_code(symbol: str, index: bool = False) -> str | None:
    """生成新浪行情接口代码（带 sh/sz 前缀）。

    指数需与股票区分：000001 既是上证指数又是平安银行，故由 index 决定
    （上证系列 000xxx -> sh，深证系列 399xxx -> sz）。
    """
    if symbol[:2].lower() in ("sh", "sz"):
        return symbol                      # 已带前缀，直接用
    code = bare_code(symbol)
    if index:
        prefix = "sh" if code.startswith("000") else "sz"
    elif is_etf(symbol):
        prefix = "sz" if code.startswith(("15", "16", "18")) else "sh"
    else:
        prefix = "sz" if code.startswith(("0", "3")) else "sh"
    return f"{prefix}{code}"


def fetch_today_sina(symbol: str, index: bool = False) -> pd.DataFrame | None:
    """新浪实时行情接口：取**当日** OHLCV，用于补历史接口的滞后。

    历史日线接口（东财/新浪）最新只到上一交易日，拿不到当天。这里直连
    hq.sinajs.cn（不经过 akshare），需带 Referer 防盗链，响应为 GBK。
    返回单行 DataFrame（date/open/close/high/low/volume）；停牌、代码不存在
    或非交易日返回 None（由调用方静默跳过）。
    """
    q = sina_quote_code(symbol, index=index)
    if q is None:
        return None
    try:
        resp = requests.get(
            f"https://hq.sinajs.cn/list={q}",
            headers={"Referer": "https://finance.sina.com.cn"},
            timeout=10,
        )
        resp.encoding = "gbk"
        body = resp.text
    except Exception as exc:  # noqa: BLE001
        print(f"[warn] 新浪当日行情接口失败: {exc}")
        return None

    # var hq_str_sh562500="名称,今开,昨收,现价,最高,最低,买一,卖一,成交量,成交额,…,日期,时间,…";
    m = re.search(r'"([^"]*)"', body)
    if not m:
        return None
    f = m.group(1).split(",")
    if len(f) < 32 or not f[1] or not f[3]:
        return None                        # 今开/现价为空 = 停牌或代码不存在
    try:
        row = {
            "date": pd.to_datetime(f[30]),
            "open": float(f[1]),
            "close": float(f[3]),
            "high": float(f[4]),
            "low": float(f[5]),
            "volume": float(f[8]),
        }
    except (ValueError, IndexError):
        return None
    if row["close"] <= 0 or row["open"] <= 0:
        return None
    return pd.DataFrame([row])


def fetch_and_cache(
    symbol: str,
    period: str,
    start: str,
    end: str,
    force: bool = False,
    ma_windows: tuple[int, ...] = (5, 10, 20),
    macd_params: tuple[int, int, int] = (12, 26, 9),
    index: bool = False,
) -> Path:
    """联网拉取行情并写入本地 CSV 缓存，返回缓存路径。"""
    target = DATA_DIR / f"{symbol}_{period}.csv"
    if target.exists() and not force:
        print(f"[skip] 已存在 {target}（如需重拉加 --force）")
        return target

    print(
        f"[fetch] {symbol}{'[指数]' if index else ''} period={period}"
        f" MA={ma_windows} MACD={macd_params} ..."
    )
    raw = fetch_raw(symbol, period, start, end, index=index)
    if raw is None or raw.empty:
        # 失败不覆盖现有缓存（to_csv 只在完整成功后执行）；
        # 区分两种成因，避免把"被限流"误报为"空数据"
        raise RuntimeError(
            "接口返回空数据：证券代码错误，或东财/新浪均被限流——"
            "请稍后重试或检查代码"
        )

    df = normalize(raw)
    # 历史日线接口最新只到上一交易日，默认补当日 K 线。
    # 同一天重复运行时覆盖末行，保证盘中刷新后指标是最新的。
    if period == "daily":
        today = fetch_today_sina(symbol, index=index)
        if today is not None:
            tdate = today["date"].iloc[0]
            if not df.empty and df["date"].iloc[-1] == tdate:
                df = pd.concat([df.iloc[:-1], today], ignore_index=True)
            elif df.empty or df["date"].iloc[-1] < tdate:
                df = pd.concat([df, today], ignore_index=True)
            print(f"[today] 补齐 {tdate.date()} 收={today['close'].iloc[0]:g}"
                  f"（实时行情，未收盘时为盘中值）")
    df = enrich(df, ma_windows, macd_params)
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    df.to_csv(target, index=False)
    # 顺带缓存证券名称供报告展示（接口失败则静默降级为只显示代码）
    name = resolve_name(symbol, index=index)
    label = f"{symbol} {name}" if name else symbol
    print(
        f"[done] {label} {len(df)} 行 -> {target}"
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
        "--index", action="store_true",
        help="证券代码为指数（如 000001=上证指数，走新浪接口，仅日线/周线）",
    )
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
    parser.add_argument(
        "--name",
        default=None,
        help="手动指定证券名称（ETF 名称接口常被反爬，可在此补全，"
             "写入 data/meta.json 供报告展示）",
    )
    args = parser.parse_args()

    if args.name:
        save_name(args.symbol, args.name)
        print(f"[name] {args.symbol} -> {args.name}（已写入 meta.json）")

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
        index=args.index,
    )


if __name__ == "__main__":
    main()
