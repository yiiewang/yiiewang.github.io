"""交互式 HTML 回测报告 v3：覆盖 analyze 八段报告的全部信息。

用法（在本目录下）：

    .venv/bin/python chart.py --symbol 562500 --period daily
    .venv/bin/python chart.py --symbol 000001 --period daily   # 上证指数

输出 reports/chart_{symbol}_{period}.html（配合 ../assets/echarts.min.js
离线打开）。相对 v2 新增：

- 顶部状态卡：市场状态判定（三条件）+ 四信号当前多空
- 摘要扩展：年化 / 夏普 / 卡尔玛 / 信号统计（入场离场分计、年均、间隔）
- 策略清单全量（23 个，下拉分组）：单指标4 + 双确认6 + 三确认3 +
  OR对偶1 + 过滤版6（ADX20/ADX25/DIF零上/底背离/周线MA/周线MACD）+
  三档+冷却1 + 切换系统2（普通/网格方向过滤）
- 三档止盈：操作列表按"轮"展示（一次买入 + 三档分批出场 + 离场原因）
- 切换系统：主图背景条带显示趋势/网格/现金模式分布，混合操作列表
- 信号质量卡：入场x零轴/底背离、离场x顶背离分组统计（随策略切换）
- 交互保留：点击列表定位 K 线区间、点击 K 线查当日持仓浮盈、
  同区间跨策略收益对比、净值曲线对比（与主图联动）
- 策略排名表：全策略（含买入持有）收益/年化/回撤/夏普/卡尔玛/胜率
  对比，综合分（五项排名均值）默认排序，点表头换列，点行切策略
- 子图区（主图同宽，左列内）：MACD（DIF/DEA/柱 + 底/顶背离标记）
  + ADX/DI（20/25 阈值）+ 净值曲线，与主图缩放/十字线联动，缺列
  自动隐藏；浮窗开关控制各图 tooltip 浮层显隐（关闭时保留十字
  准线；净值默认关闭）；右侧列表拉伸至与主图+子图总高等高，仅
  纵向滚动（长文本折行、禁横向溢出）
"""

from __future__ import annotations

import argparse
import json
import re
from itertools import combinations
from pathlib import Path

import pandas as pd

from analyze import (
    DATA_DIR,
    PERIOD_NAMES,
    _has_transitive_redundancy,
    build_all_masks,
    build_filters,
    buy_and_hold,
    collect_current_state,
    collect_density_triggers,
    density_hold,
    higher_permission,
    load,
    ma_columns,
    perf_metrics,
    regime_mask,
    signal_bulls,
    state_backtest,
    symbol_label,
)
from grid import switch_backtest
from ladder import ladder_backtest

HERE = Path(__file__).resolve().parent
TEMPLATE_PATH = HERE / "templates" / "report.html"

# 报告数据结构的契约版本。数据层与模板层只通过它声明兼容性：
# 结构发生不兼容变更时递增，避免字段改动静默破坏渲染。
REPORT_SCHEMA_VERSION = 1


def _f2(v: float | None) -> str:
    """perf 指标的安全格式化（零波动策略的 sharpe/calmar 可能为 None）。"""
    return f"{v:.2f}" if v is not None else "--"


def state_payload(st: dict) -> dict:
    """把 collect_current_state 的结果转成前端渲染用的 {k, v, tone} 行。

    只做展示格式转换——判定逻辑全部留在 analyze.collect_current_state，
    保证网页状态面板与终端「当前状态」段同源同口径。
    """

    def g(x: float) -> str:
        return f"{x:.4g}"

    signals = [{
        "k": s["name"],
        "v": (f"{s['fast_label']}({g(s['fast'])}) {'>' if s['bull'] else '<'} "
              f"{s['slow_label']}({g(s['slow'])}) "
              f"{'多头' if s['bull'] else '空头'} 已{s['run']}根"),
        "tone": "bull" if s["bull"] else "bear",
    } for s in st["signals"]]

    combos = [{
        "k": c["label"],
        "v": f"{'多头' if c['bull'] else '空头'} {c['members_bull']}/{c['members']} 已{c['run']}根",
        "tone": "bull" if c["bull"] else "bear",
    } for c in st["combos"]]

    filters = [{
        "k": f["label"],
        "v": f"{'通过' if f['ok'] else '不通过'} {f['detail']}".strip(),
        "tone": "bull" if f["ok"] else "",
    } for f in st["filters"]]

    env = []
    for e in st["env"]:
        extra = e.get("extra", "")
        env.append({
            "k": e["label"],
            "v": e["value"] + (f" {extra}" if extra else ""),
            "tone": "bear" if extra == "跌破" else "",
        })

    divs = [{
        "k": d["label"],
        "v": d["text"],
        "tone": "warn" if d["has"] else "",
    } for d in st["divs"]]

    # 顶部状态条摘要：只挑一眼要看的关键项，细节留在右侧面板
    chips: list[dict] = []

    def env_of(label: str) -> dict | None:
        return next((e for e in st["env"] if e["label"] == label), None)

    reg = env_of("regime")
    if reg:
        tone = "bull" if "趋势↑" in reg["value"] else "bear" if "趋势↓" in reg["value"] else ""
        chips.append({"k": "regime", "v": reg["value"], "tone": tone})
    if st["combos"]:
        c = st["combos"][0]
        chips.append({
            "k": "双确认",
            "v": f"{'多头' if c['bull'] else '空头'} {c['run']}根",
            "tone": "bull" if c["bull"] else "bear",
        })
    if st["filters"]:
        ok = sum(1 for f in st["filters"] if f["ok"])
        chips.append({
            "k": "过滤器", "v": f"{ok}/{len(st['filters'])} 通过",
            "tone": "bull" if ok == len(st["filters"]) else "",
        })
    life = next((e for e in st["env"] if e["label"].startswith("生命线")), None)
    if life and life.get("extra"):
        chips.append({
            "k": "生命线", "v": life["extra"],
            "tone": "bull" if life["extra"] == "站稳" else "bear",
        })
    dd = env_of("回撤")
    if dd:
        chips.append({"k": "回撤", "v": dd["value"], "tone": "bear"})
    for d in st["divs"]:
        if d["has"]:
            txt = d["text"]
            chips.append({
                "k": d["label"],
                "v": txt.split("（")[-1].rstrip("）") if "（" in txt else txt,
                "tone": "warn",
            })

    return {
        "date": st["date"], "close": st["close"],
        "signals": signals, "combos": combos,
        "filters": filters, "env": env, "divs": divs,
        "chips": chips,
    }


def _quality_line(df: pd.DataFrame, hold: pd.Series) -> str:
    """策略级信号质量分组：入场x零轴/底背离、离场x顶背离（一行汇总）。"""
    if not {"dif_pos", "bull_div", "bear_div"}.issubset(df.columns):
        return ""
    closes = df["close"].astype(float).tolist()
    n = len(df)
    prev = hold.shift(1).fillna(False).astype(bool)
    entries = list(df.index[hold & ~prev])
    exits = list(df.index[~hold & prev])

    def avg(idxs, horizon=30):
        if not idxs:
            return "0次"
        rets = [closes[min(i + horizon, n - 1)] / closes[i] - 1 for i in idxs]
        good = sum(1 for r in rets if r > 0)
        return f"{len(idxs)}次{sum(rets)/len(rets):+.1%}/对{good/len(idxs):.0%}"

    def has(dv, t, lb=10):
        return bool(dv.iloc[max(0, t - lb): t + 1].any())

    dp, bd_, brd = df["dif_pos"], df["bull_div"], df["bear_div"]
    up = [t for t in entries if dp.iloc[t]]
    dn = [t for t in entries if not dp.iloc[t]]
    wd = [t for t in entries if has(bd_, t)]
    nd = [t for t in entries if not has(bd_, t)]
    wb = [t for t in exits if has(brd, t)]
    nb = [t for t in exits if not has(brd, t)]
    return (
        f"<b>信号质量</b>（事件后30根）　入场×零轴: 零上{avg(up)} vs 零下{avg(dn)}"
        f"　入场×底背离: 有{avg(wd)} vs 无{avg(nd)}"
        f"　离场×顶背离: 有{avg(wb)} vs 无{avg(nb)}"
    )


def _mask_payload(df: pd.DataFrame, hold: pd.Series, quality: bool = True) -> dict:
    """掩码型策略的完整数据（v2 结构 + 质量行）。"""
    should = hold.shift(1).fillna(False).astype(bool)
    opens = df["open"].astype(float).tolist()
    closes = df["close"].astype(float).tolist()
    dates = [d.strftime("%Y-%m-%d") for d in df["date"]]
    n = len(df)
    s = should.tolist()
    buys, sells, trade_list, hold_costs = [], [], [], [None] * n
    entry = None

    def close_trade(exit_i, exit_p):
        bi, bp = entry
        seg = closes[bi: exit_i + 1]
        trade_list.append({
            "bi": bi, "si": exit_i, "bd": dates[bi], "sd": dates[exit_i],
            "bp": round(bp, 4), "sp": round(exit_p, 4),
            "ret": round((exit_p / bp - 1) * 100, 2),
            "runup": round((max(seg) / bp - 1) * 100, 2),
            "dd": round((min(seg) / bp - 1) * 100, 2),
            "days": exit_i - bi,
        })

    for i in range(n):
        if s[i] and entry is None:
            entry = (i, opens[i])
            buys.append([dates[i], round(opens[i], 4)])
        elif not s[i] and entry is not None:
            sells.append([dates[i], round(opens[i], 4)])
            close_trade(i, opens[i])
            entry = None
        if entry is not None:
            hold_costs[i] = round(entry[1], 4)
    if entry is not None:
        close_trade(n - 1, closes[-1])

    r = state_backtest(df, hold)
    m = perf_metrics(r["equity_curve"])
    years = max(n / 252, 1e-9)
    wr = f"{r['win_rate']:.1%}" if r["win_rate"] is not None else "--"
    return {
        "kind": "mask",
        "buys": buys, "sells": sells,
        "tradeList": trade_list, "holdCosts": hold_costs,
        "equity": [round(v, 4) for v in r["equity_curve"].iloc[1:].tolist()],
        "metrics": {
            "ret": round(r["total_return"] * 100, 2),
            "annual": round(m["annual"] * 100, 2),
            "dd": round(r["max_dd"] * 100, 2),
            "sharpe": m["sharpe"], "calmar": m["calmar"],
            "trades": r["trades"],
            "win": None if r["win_rate"] is None else round(r["win_rate"] * 100, 1),
        },
        "desc": (f"{r['trades']} 笔　胜率 {wr}　收益 <b>{r['total_return']:+.1%}</b>"
                 f"　回撤 {r['max_dd']:.1%}　年化 {m['annual']:+.1%}"
                 f"　夏普 {_f2(m['sharpe'])}　卡尔玛 {_f2(m['calmar'])}"
                 f"　年均入场 {len(buys)/years:.1f} 次"),
        "quality": _quality_line(df, hold) if quality else "",
    }


def _ladder_payload(df: pd.DataFrame, hold: pd.Series) -> dict:
    """三档止盈：按轮展示（一次买入 + 三档分批出场）。"""
    r = ladder_backtest(df, hold, exit_on_signal=True, cooldown=True)
    dates = [d.strftime("%Y-%m-%d") for d in df["date"]]
    opens = df["open"].astype(float).tolist()
    rounds = []
    buys, sells = [], []
    for t in r["trades"]:
        tiers = " / ".join(
            f"{x:.4g}" if x is not None else "-" for x in t["exits"]
        )
        sd = dates[max(i for i in t["exit_is"] if i is not None)]
        rounds.append({
            "bi": t["entry_i"], "maxSi": max(i for i in t["exit_is"] if i is not None),
            "bd": dates[t["entry_i"]], "sd": sd, "bp": round(t["entry"], 4),
            "tiers": tiers, "ret": round((t["factor"] - 1) * 100, 2),
            "reason": t["reason"],
        })
        buys.append([dates[t["entry_i"]], round(t["entry"], 4)])
        for x, i in zip(t["exits"], t["exit_is"]):
            if x is not None and i is not None:
                sells.append([dates[i], round(x, 4)])
    m = perf_metrics(r["equity_curve"])
    wr = f"{r['win_rate']:.1%}" if r["win_rate"] is not None else "--"
    return {
        "kind": "ladder",
        "buys": buys, "sells": sells, "rounds": rounds,
        "equity": [round(v, 4) for v in r["equity_curve"].iloc[1:].tolist()],
        "metrics": {
            "ret": round(r["total_return"] * 100, 2),
            "annual": round(m["annual"] * 100, 2),
            "dd": round(r["max_dd"] * 100, 2),
            "sharpe": m["sharpe"], "calmar": m["calmar"],
            "trades": r["count"],
            "win": None if r["win_rate"] is None else round(r["win_rate"] * 100, 1),
        },
        "desc": (f"{r['count']} 轮　胜率 {wr}　收益 <b>{r['total_return']:+.1%}</b>"
                 f"　回撤 {r['max_dd']:.1%}　年化 {m['annual']:+.1%}"
                 f"　夏普 {_f2(m['sharpe'])}　卡尔玛 {_f2(m['calmar'])}"),
        "quality": "",
    }


def _switch_payload(df: pd.DataFrame, hold: pd.Series, rmask: pd.Series) -> dict:
    """切换系统：模式条带 + 趋势笔/网格轮混合列表。"""
    r = switch_backtest(df, hold, rmask)
    dates = [d.strftime("%Y-%m-%d") for d in df["date"]]
    # 模式段合并（连续同模式 -> markArea 段）
    areas, start = [], 0
    modes = r["mode_series"]
    for i in range(1, len(modes) + 1):
        if i == len(modes) or modes[i] != modes[start]:
            color = {"trend": "rgba(194,53,49,.08)", "grid": "rgba(46,139,87,.10)",
                     "cash": "rgba(160,160,160,.08)"}[modes[start]]
            areas.append([start, i - 1, color])
            start = i
    events = []
    for t in r["trend_trade_list"]:
        events.append({"k": "t", "bi": t["bi"], "ei": t["si"], "bd": dates[t["bi"]],
                       "ed": dates[t["si"]], "ret": t["ret"], "days": t["days"],
                       "info": "单笔管理", "si": t["si"]})
    for g in r["grid_round_list"]:
        events.append({"k": "g", "bi": g["si"], "ei": g["ei"], "bd": dates[g["si"]],
                       "ed": dates[g["ei"]], "ret": g["ret"],
                       "days": g["ei"] - g["si"], "info": f"{len(g['fills'])} 次成交",
                       "si": g["ei"]})
    events.sort(key=lambda e: e["bi"])
    m = perf_metrics(r["equity_curve"])
    days = r["mode_days"]
    total = sum(days.values())
    return {
        "kind": "switch", "areas": areas, "events": events,
        "equity": [round(v, 4) for v in r["equity_curve"].iloc[1:].tolist()],
        "metrics": {
            "ret": round(r["total_return"] * 100, 2),
            "annual": round(m["annual"] * 100, 2),
            "dd": round(r["max_dd"] * 100, 2),
            "sharpe": m["sharpe"], "calmar": m["calmar"],
            "trades": len(events), "win": None,
        },
        "desc": (f"趋势{days['trend']/total:.0%}/网格{days['grid']/total:.0%}/现金{days['cash']/total:.0%}"
                 f"　收益 <b>{r['total_return']:+.1%}</b>　回撤 {r['max_dd']:.1%}"
                 f"　年化 {m['annual']:+.1%}　夏普 {_f2(m['sharpe'])}　卡尔玛 {_f2(m['calmar'])}"),
        "quality": "",
    }


def compute_density(
    df: pd.DataFrame, bulls: dict[str, pd.Series], combo: pd.Series,
    rmask: pd.Series | None, higher_df: pd.DataFrame | None,
    n_days: int = 5,
) -> dict:
    """计算最近 N 天内各策略的信号触发事件，供前端状态条与主图标记。

    返回 {buys, sells, warns, rating, triggers}，其中 triggers 是
    [{date, dir, name}] 列表（按日期倒序），buys/sells/warns 是计数。
    """
    triggers = collect_density_triggers(df, bulls, combo, rmask, higher_df, n_days)

    buys = sum(1 for t in triggers if t[1] == "买入")
    sells = sum(1 for t in triggers if t[1] == "卖出")
    warns = sum(1 for t in triggers if t[1] == "预警")
    total = buys + sells
    rating = ("高度密集" if total >= 5 else "中等密集" if total >= 3
              else "零星" if total >= 1 else "无信号")
    triggers.sort(key=lambda t: t[2], reverse=True)
    return {
        "buys": buys, "sells": sells, "warns": warns,
        "rating": rating, "nDays": n_days,
        "triggers": [{"date": d.strftime("%Y-%m-%d"), "dir": direction, "name": name}
                      for name, direction, d in triggers],
    }


def render(payload: dict, out: Path) -> Path:
    """用 templates/report.html 渲染数据，产出单文件 HTML。

    数据与展示分离：payload 只依赖 schema，模板只依赖三个占位符
    （__DATA__ / __TITLE__ / __ECHARTS__）——换外观只需替换模板文件，
    Python 侧不动。
    echarts 内嵌进 HTML：单文件零外部引用，file:// 双击、IDE 预览、
    http 服务任何打开方式都稳健（相对引用在部分预览环境下会 404）。
    """
    if not TEMPLATE_PATH.exists():
        raise SystemExit(f"缺少模板文件 {TEMPLATE_PATH}")
    template = TEMPLATE_PATH.read_text(encoding="utf-8")
    echarts_js = (HERE / "assets" / "echarts.min.js").read_text(encoding="utf-8")
    html = (
        template
        .replace("__ECHARTS__", echarts_js)
        .replace("__DATA__", json.dumps(payload, ensure_ascii=False))
        .replace("__TITLE__", f"{payload['symbolLabel']} 策略回测报告")
    )
    out.write_text(html, encoding="utf-8")
    return out


def make_chart(symbol: str, period: str) -> Path:
    df = load(symbol, period)
    bulls = signal_bulls(df)
    cols = ma_columns(df)
    pair = f"MA{cols[0][2:]}x{cols[-1][2:]}"
    rmask = regime_mask(df)
    higher_df = None
    wp = DATA_DIR.parent / "data" / f"{symbol}_weekly.csv"
    if wp.exists():
        try:
            higher_df = load(symbol, "weekly")
        except Exception:
            higher_df = None

    strategies: dict[str, dict] = {}
    groups: list[dict] = []

    # 单指标
    for name, hold in bulls.items():
        strategies[name] = _mask_payload(df, hold)
    groups.append({"label": "单指标", "names": list(bulls)})
    combo = bulls[pair] & bulls["MACD"] if "MACD" in bulls else None

    # 双确认 / 三确认 / OR（枚举，跳过传递性冗余）
    names = list(bulls)

    for size, label in ((2, "双确认"), (3, "三确认")):
        members = []
        for c in combinations(names, size):
            if size == 3 and _has_transitive_redundancy(c):
                continue
            n = "+".join(c)
            h = bulls[c[0]]
            for m in c[1:]:
                h = h & bulls[m]
            strategies[n] = _mask_payload(df, h, quality=False)
            members.append(n)
        groups.append({"label": label, "names": members})
    or_hold = None
    for h in bulls.values():
        or_hold = h if or_hold is None else or_hold | h
    strategies["任一确认(OR)"] = _mask_payload(df, or_hold, quality=False)
    groups.append({"label": "对偶", "names": ["任一确认(OR)"]})

    # 过滤版（双确认 x 6 种过滤器）
    fmembers = []
    for label, fmask in build_filters(df, higher_df, pair):
        n = f"{pair}+MACD+{label}"
        strategies[n] = _mask_payload(df, combo & fmask, quality=False)
        fmembers.append(n)
    groups.append({"label": "过滤版", "names": fmembers})

    # 三档 + 切换
    strategies["三档+冷却协议"] = _ladder_payload(df, combo)
    groups.append({"label": "仓位管理", "names": ["三档+冷却协议"]})
    if rmask is not None:
        strategies["切换系统"] = _switch_payload(df, combo, rmask)
        sw_names = ["切换系统"]
        if higher_df is not None:
            wperm = higher_permission(df, higher_df, "MACD")
            if wperm is not None:
                strategies["切换+网格方向过滤"] = _switch_payload(df, combo, rmask & wperm)
                sw_names.append("切换+网格方向过滤")
        groups.append({"label": "切换系统", "names": sw_names})

    # 信号密度策略
    d_hold = density_hold(df, higher_df, n_days=5, threshold=3)
    if d_hold.any():
        strategies["信号密度(5天/3)"] = _mask_payload(df, d_hold, quality=False)
        groups.append({"label": "密度策略", "names": ["信号密度(5天/3)"]})

    # 顶部状态卡
    cards = []
    last = df.iloc[-1]
    if rmask is not None:
        score = int(rmask.iloc[-1])
        cards.append(
            f"<b>市场状态</b> {last['date'].date()}　close={last['close']:g}"
            f"　ADX={last['adx']:.1f}　粘合={last['cohesion']:.1%}"
            f"　震荡条件 {score}/3"
            f"（{'+' if last['pdi'] > last['mdi'] else '-'}DI"
            f" {'上' if last['pdi'] > last['mdi'] else '下'}行占优）"
        )
    sig_lines = []
    for name in bulls:
        sig_lines.append(f"{name}{'多' if bulls[name].iloc[-1] else '空'}")
    cards.append("<b>当前信号</b> " + " ".join(sig_lines))

    # 状态摘要并入顶部状态卡（与「市场状态」同 UI、同一行）；
    # regime 不重复加——市场状态卡已表达同一结论
    st = state_payload(collect_current_state(df, higher_df))
    _tone = {"bull": "#c23531", "bear": "#2e8b57", "warn": "#a06d20"}
    for c in st["chips"]:
        if c["k"] == "regime":
            continue
        cards.append(
            f'<b>{c["k"]}</b> <span style="color:{_tone.get(c["tone"], "#555")}">{c["v"]}</span>'
        )

    closes = df["close"].astype(float).tolist()
    hold_m = perf_metrics(pd.Series(closes, dtype=float) / closes[0])
    cats = [d.strftime("%Y-%m-%d") for d in df["date"]]

    # 指标子图数据（缓存缺列时前端自动隐藏对应栏）
    def _ind_num(col: str) -> list:
        return [None if pd.isna(v) else round(float(v), 4) for v in df[col]]

    ind: dict = {}
    if {"dif", "dea"}.issubset(df.columns):
        dif_s, dea_s = _ind_num("dif"), _ind_num("dea")
        ind.update(dif=dif_s, dea=dea_s,
                   hist=[None if a is None or b is None else round(a - b, 4)
                         for a, b in zip(dif_s, dea_s)])
        if "bull_div" in df.columns:
            flags = df["bull_div"].fillna(False).astype(bool)
            ind["bull"] = [[d, v] for d, v, m in zip(cats, dif_s, flags) if m and v is not None]
        if "bear_div" in df.columns:
            flags = df["bear_div"].fillna(False).astype(bool)
            ind["bear"] = [[d, v] for d, v, m in zip(cats, dif_s, flags) if m and v is not None]
    if {"adx", "pdi", "mdi"}.issubset(df.columns):
        ind.update(adx=_ind_num("adx"), pdi=_ind_num("pdi"), mdi=_ind_num("mdi"))

    payload = {
        "schemaVersion": REPORT_SCHEMA_VERSION,
        "symbol": symbol, "period": period,
        "symbolLabel": symbol_label(symbol),
        "periodName": PERIOD_NAMES.get(period, period),
        "span": f"{df['date'].iloc[0].date()} ~ {df['date'].iloc[-1].date()}",
        "categories": cats,
        "kline": [
            [round(float(r.open), 4), round(float(r.close), 4),
             round(float(r.low), 4), round(float(r.high), 4)]
            for r in df.itertuples()
        ],
        "holdEquity": [round(c / closes[0], 4) for c in closes],
        "holdMetrics": {
            "ret": round(hold_m["total"] * 100, 2),
            "annual": round(hold_m["annual"] * 100, 2),
            "dd": round(hold_m["max_dd"] * 100, 2),
            "sharpe": hold_m["sharpe"], "calmar": hold_m["calmar"],
            "trades": 1, "win": None,
        },
        "cards": cards,
        "groups": groups,
        "strategies": strategies,
        "ind": ind,
        "density": compute_density(df, bulls, combo, rmask, higher_df),
        "state": st,
    }
    for col in ("ma5", "ma10", "ma20"):
        payload[col] = [None if pd.isna(v) else round(float(v), 4) for v in df[col]]

    # ① 数据独立落盘：与展示解耦，其他消费方可直接读 JSON 而不必解析 HTML
    data_dir = HERE / "reports" / "data"
    data_dir.mkdir(parents=True, exist_ok=True)
    data_file = data_dir / f"{symbol}_{period}.json"
    data_file.write_text(
        json.dumps(payload, ensure_ascii=False, indent=1), encoding="utf-8"
    )

    # ② 用模板渲染：换外观只替换 templates/report.html，Python 侧不动
    out_dir = HERE / "reports"
    out_dir.mkdir(parents=True, exist_ok=True)
    out = out_dir / f"chart_{symbol}_{period}.html"
    render(payload, out)
    print(
        f"[chart] {out.relative_to(HERE)}（{len(df)} 根K线，"
        f"{len(strategies)} 个策略，{len(groups)} 组）"
    )
    print(f"[data] {data_file.relative_to(HERE)}（schema v{REPORT_SCHEMA_VERSION}）")
    return out


def main() -> None:
    parser = argparse.ArgumentParser(description="生成交互式回测报告 HTML")
    parser.add_argument("--symbol", default="562500", help="证券代码")
    parser.add_argument("--period", default="daily", help="daily/weekly/1/5/...")
    args = parser.parse_args()
    make_chart(args.symbol, args.period)


if __name__ == "__main__":
    main()
