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
"""

from __future__ import annotations

import argparse
import json
from itertools import combinations
from pathlib import Path

import pandas as pd

from analyze import (
    DATA_DIR,
    buy_and_hold,
    higher_permission,
    load,
    ma_columns,
    perf_metrics,
    regime_mask,
    signal_bulls,
    state_backtest,
)
from grid import switch_backtest
from ladder import ladder_backtest

HERE = Path(__file__).resolve().parent


def _f2(v: float | None) -> str:
    """perf 指标的安全格式化（零波动策略的 sharpe/calmar 可能为 None）。"""
    return f"{v:.2f}" if v is not None else "--"

HTML_TEMPLATE = """<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="utf-8">
<title>{title}</title>
<style>
  * { box-sizing: border-box; }
  body { margin: 0; padding: 12px; font: 14px/1.6 -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif; background: #fafafa; }
  .cards { display: flex; gap: 10px; flex-wrap: wrap; margin-bottom: 8px; }
  .card { background: #fff; border: 1px solid #e5e5e5; padding: 6px 12px; font-size: 12.5px; color: #444; border-radius: 3px; }
  .card b { color: #c23531; }
  .bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; margin-bottom: 8px; }
  select { font-size: 14px; padding: 3px 6px; max-width: 340px; }
  #summary { color: #333; font-size: 13px; }
  #summary b { color: #c23531; }
  .main { display: grid; grid-template-columns: 1fr 360px; gap: 10px; }
  #mainChart { width: 100%; height: 500px; background: #fff; border: 1px solid #eee; }
  .side { display: flex; flex-direction: column; gap: 6px; }
  #dayInfo { min-height: 18px; color: #666; font-size: 12.5px; }
  #quality { background: #eef6ff; border: 1px solid #c5dcf5; padding: 6px 10px; font-size: 12px; color: #345; }
  #detail { background: #fff8e6; border: 1px solid #f0e0b0; padding: 6px 10px; font-size: 12.5px; display: none; }
  #trades { background: #fff; border: 1px solid #eee; height: 330px; overflow-y: auto; }
  .t-head, .t-row { display: grid; grid-template-columns: 30px 1fr 96px 58px 40px; gap: 4px; padding: 3px 8px; font-size: 12px; }
  .t-head { position: sticky; top: 0; background: #f0f0f0; font-weight: 600; z-index: 2; }
  .t-row { border-top: 1px solid #f5f5f5; cursor: pointer; }
  .t-row:hover { background: #fdf3e0; }
  .t-row.active { background: #ffe9b8; }
  .tag { font-size: 10px; padding: 0 3px; border-radius: 2px; color: #fff; margin-right: 3px; }
  .tag.t { background: #c23531; } .tag.g { background: #2e8b57; }
  #eqChart { width: 100%; height: 240px; background: #fff; border: 1px solid #eee; margin-top: 10px; }
  .tip { color: #888; font-size: 12px; margin-top: 5px; }
  .pos { color: #c23531; } .neg { color: #2e8b57; }
</style>
</head>
<body>
<div class="cards" id="cards"></div>
<div class="bar">
  <label>策略：<select id="strategy"></select></label>
  <span id="summary"></span>
</div>
<div class="main">
  <div>
    <div id="mainChart"></div>
    <div id="dayInfo" style="margin-top:4px">点击K线任意一天可查看该日持仓与浮盈</div>
  </div>
  <div class="side">
    <div id="detail"></div>
    <div id="quality"></div>
    <div id="trades"></div>
  </div>
</div>
<div id="eqChart"></div>
<p class="tip">▲ 买入 = 转多的次日开盘执行价；▼ 卖出 = 转空的次日开盘（与回测同语义）。点击列表定位 K 线区间并显示同期跨策略对比；切换系统选中时主图背景为模式条带（红=趋势/绿=网格/灰=现金）；底部净值图与主图联动。</p>
<script>__ECHARTS__</script>
<script>
const RAW = __DATA__;

const mainChart = echarts.init(document.getElementById('mainChart'));
const eqChart = echarts.init(document.getElementById('eqChart'));
echarts.connect([mainChart, eqChart]);
const N = RAW.categories.length;
const zoomStart = Math.max(0, 100 * (1 - 250 / N));
const cat = { type: 'category', data: RAW.categories, axisLabel: { formatter: v => v.slice(2) } };

// 顶部状态卡
document.getElementById('cards').innerHTML =
  RAW.cards.map(c => '<div class="card">' + c + '</div>').join('');

mainChart.setOption({
  animation: false,
  tooltip: { trigger: 'axis', axisPointer: { type: 'cross' } },
  legend: { top: 4, data: ['K线', 'MA5', 'MA10', 'MA20'] },
  grid: { left: 60, right: 16, top: 34, bottom: 52 },
  xAxis: cat,
  yAxis: { scale: true, splitArea: { show: true } },
  dataZoom: [
    { type: 'inside', start: zoomStart, end: 100 },
    { type: 'slider', start: zoomStart, end: 100, height: 18, bottom: 10 },
  ],
  series: [
    { name: 'K线', type: 'candlestick', data: RAW.kline, markArea: { silent: true, data: [] },
      itemStyle: { color: '#ef232a', color0: '#14b143', borderColor: '#ef232a', borderColor0: '#14b143' } },
    { name: 'MA5', type: 'line', data: RAW.ma5, showSymbol: false, lineStyle: { width: 1, opacity: .8 } },
    { name: 'MA10', type: 'line', data: RAW.ma10, showSymbol: false, lineStyle: { width: 1, opacity: .8 } },
    { name: 'MA20', type: 'line', data: RAW.ma20, showSymbol: false, lineStyle: { width: 1.2, opacity: .9 } },
    { id: 'buy', name: '买入', type: 'scatter', data: [], symbol: 'triangle', symbolSize: 10, z: 5,
      itemStyle: { color: '#e6a23c', borderColor: '#fff', borderWidth: .5 } },
    { id: 'sell', name: '卖出', type: 'scatter', data: [], symbol: 'triangle', symbolRotate: 180, symbolSize: 10, z: 5,
      itemStyle: { color: '#6c5ce7', borderColor: '#fff', borderWidth: .5 } },
    { id: 'fbuy', type: 'scatter', data: [], symbol: 'triangle', symbolSize: 20, z: 6,
      itemStyle: { color: '#ff8c00', borderColor: '#333', borderWidth: 1 } },
    { id: 'fsell', type: 'scatter', data: [], symbol: 'triangle', symbolRotate: 180, symbolSize: 20, z: 6,
      itemStyle: { color: '#4834d4', borderColor: '#333', borderWidth: 1 } },
  ],
});

const eqNames = Object.keys(RAW.strategies).concat(['买入持有']);
eqChart.setOption({
  animation: false,
  tooltip: { trigger: 'axis' },
  legend: { top: 2, type: 'scroll' },
  grid: { left: 60, right: 16, top: 30, bottom: 24 },
  xAxis: cat,
  yAxis: { type: 'log', scale: true, axisLabel: { formatter: v => v.toFixed(2) } },
  dataZoom: [{ type: 'inside', start: zoomStart, end: 100 }],
  series: eqNames.map(n => ({
    name: n, type: 'line', showSymbol: false,
    lineStyle: { width: n.includes('双确认') || n === '买入持有' ? 1.8 : 1, opacity: .85 },
    data: n === '买入持有' ? RAW.holdEquity : RAW.strategies[n].equity,
  })),
});

const select = document.getElementById('strategy');
const summary = document.getElementById('summary');
const dayInfo = document.getElementById('dayInfo');
const detail = document.getElementById('detail');
const quality = document.getElementById('quality');
const tradesBox = document.getElementById('trades');
let current = null;

const pct = v => (v >= 0 ? '+' : '') + v.toFixed(2) + '%';
const cls = v => v >= 0 ? 'pos' : 'neg';
const md = d => d.slice(5);

RAW.groups.forEach(g => {
  const og = document.createElement('optgroup');
  og.label = g.label;
  g.names.forEach(n => { const o = document.createElement('option'); o.value = n; o.textContent = n; og.appendChild(o); });
  select.appendChild(og);
});

function head(cols) {
  tradesBox.innerHTML = '<div class="t-head">' + cols.map(c => '<span>' + c + '</span>').join('') + '</div>';
}
function row(cells, bi, si, click) {
  const r = document.createElement('div');
  r.className = 't-row';
  r.innerHTML = cells.map(c => '<span>' + c + '</span>').join('');
  r.onclick = () => {
    document.querySelectorAll('.t-row.active').forEach(x => x.classList.remove('active'));
    r.classList.add('active');
    focusRange(bi, si, click);
  };
  tradesBox.appendChild(r);
}

function focusRange(bi, si, info) {
  const pad = 30;
  mainChart.dispatchAction({ type: 'dataZoom',
    start: Math.max(0, (bi - pad) * 100 / N), end: Math.min(100, (si + pad) * 100 / N) });
  const cmp = Object.entries(RAW.strategies)
    .map(([n, st]) => { const r = (st.equity[si] / st.equity[bi] - 1) * 100;
      return '<span class="' + cls(r) + '" style="margin-right:6px">' + n + ' ' + pct(r) + '</span>'; })
    .join('');
  const hr = (RAW.holdEquity[si] / RAW.holdEquity[bi] - 1) * 100;
  detail.style.display = 'block';
  detail.innerHTML = info + '<br><span style="color:#555">同区间: ' + cmp
    + '<span class="' + cls(hr) + '">持有 ' + pct(hr) + '</span></span>';
}

function markPoints(bd, bp, sd, sp) {
  mainChart.setOption({ series: [
    { id: 'fbuy', data: bd == null ? [] : [[bd, bp]] },
    { id: 'fsell', data: sd == null ? [] : [[sd, sp]] },
  ]});
}

function applyStrategy(name) {
  current = name;
  select.value = name;
  const s = RAW.strategies[name];
  mainChart.setOption({ series: [
    { id: 'buy', data: s.buys || [] },
    { id: 'sell', data: s.sells || [] },
    { id: 'fbuy', data: [] }, { id: 'fsell', data: [] },
    { name: 'K线', markArea: { silent: true, data: (s.areas || []).map(a => [
      { xAxis: RAW.categories[a[0]], itemStyle: { color: a[2] } },
      { xAxis: RAW.categories[a[1]] }]) } },
  ]});
  summary.innerHTML = '<b>' + name + '</b>　' + (s.desc || '');
  quality.innerHTML = s.quality || '';
  detail.style.display = 'none';

  if (s.kind === 'ladder') {
    head(['#', '买入→离场', '三档出场', '轮收益', '原因']);
    s.rounds.forEach((t, i) => row(
      [i + 1, md(t.bd) + '→' + md(t.sd), t.tiers, '<b class="' + cls(t.ret) + '">' + pct(t.ret) + '</b>', t.reason],
      t.bi, t.maxSi, '<b>本轮回合 ' + pct(t.ret) + '</b>（' + t.bd + ' ' + t.bp + ' → ' + t.sd + '）　三档: ' + t.tiers + '　离场原因: ' + t.reason));
  } else if (s.kind === 'switch') {
    head(['#', '类型 区间', '明细', '收益', '天数']);
    s.events.forEach((e, i) => row(
      [i + 1, (e.k === 't' ? '<span class="tag t">趋势</span>' : '<span class="tag g">网格</span>') + md(e.bd) + '→' + md(e.ed),
       e.info, '<b class="' + cls(e.ret) + '">' + pct(e.ret) + '</b>', e.days + 'd'],
      e.bi, e.ei, '<b>' + (e.k === 't' ? '趋势笔' : '网格轮') + ' ' + pct(e.ret) + '</b>（' + e.bd + ' → ' + e.ed + '）　' + e.info));
  } else {
    head(['#', '买入 → 卖出', '价格', '收益', '天数']);
    s.tradeList.forEach((t, i) => row(
      [i + 1, md(t.bd) + ' → ' + md(t.sd), t.bp + '→' + t.sp,
       '<b class="' + cls(t.ret) + '">' + pct(t.ret) + '</b>', t.days + 'd'],
      t.bi, t.si, '<b>本笔 ' + pct(t.ret) + '</b>（' + t.bd + ' ' + t.bp + ' → ' + t.sd + ' ' + t.sp + '，' + t.days + ' 天）　最高浮盈 ' + pct(t.runup) + '　最深浮亏 ' + pct(t.dd)));
    markPoints(s.tradeList.length ? null : null, 0, null, 0);
  }
}

mainChart.on('click', params => {
  if (params.seriesType !== 'candlestick') return;
  const s = RAW.strategies[current];
  const i = params.dataIndex, c = params.data[1];
  if (s.kind !== 'mask') { dayInfo.textContent = RAW.categories[i] + '：close=' + c; return; }
  const cost = s.holdCosts[i];
  dayInfo.innerHTML = cost == null
    ? RAW.categories[i] + '：空仓（close=' + c + '）'
    : RAW.categories[i] + '：持仓 成本 ' + cost + '　close=' + c
      + '　浮动盈亏 <b class="' + cls((c / cost - 1) * 100) + '">' + pct((c / cost - 1) * 100) + '</b>';
});

select.addEventListener('change', () => applyStrategy(select.value));
applyStrategy('三档+冷却协议');
window.addEventListener('resize', () => { mainChart.resize(); eqChart.resize(); });
</script>
</body>
</html>
"""


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
        "desc": (f"趋势{days['trend']/total:.0%}/网格{days['grid']/total:.0%}/现金{days['cash']/total:.0%}"
                 f"　收益 <b>{r['total_return']:+.1%}</b>　回撤 {r['max_dd']:.1%}"
                 f"　年化 {m['annual']:+.1%}　夏普 {_f2(m['sharpe'])}　卡尔玛 {_f2(m['calmar'])}"),
        "quality": "",
    }


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
    dup = lambda c: False
    import re as _re

    def transitive(c):
        pairs = []
        for x in c:
            m = _re.fullmatch(r"MA(\d+)x(\d+)", x)
            if m:
                pairs.append((int(m.group(1)), int(m.group(2))))
        ps = set(pairs)
        return any(b == c2 and (a, d) in ps for a, b in pairs for c2, d in pairs)

    for size, label in ((2, "双确认"), (3, "三确认")):
        members = []
        for c in combinations(names, size):
            if size == 3 and transitive(c):
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
    filters = []
    if "adx" in df.columns:
        filters.append(("ADX>=20", df["adx"] >= 20))
        filters.append(("ADX>=25", df["adx"] >= 25))
    if "dif_pos" in df.columns:
        filters.append(("DIF零上", df["dif_pos"]))
    if "bull_div" in df.columns:
        filters.append(("底背离前兆", df["bull_div"].astype(bool).rolling(10, min_periods=1).max().astype(bool)))
    if higher_df is not None:
        for src, lab in ((pair, f"周线{pair}"), ("MACD", "周线MACD")):
            mask = higher_permission(df, higher_df, src)
            if mask is not None:
                filters.append((f"{lab}多头", mask))
    fmembers = []
    for label, fmask in filters:
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
    for col, name in signal_bulls_names(df, bulls):
        sig_lines.append(f"{name}{'多' if bulls[name].iloc[-1] else '空'}")
    cards.append("<b>当前信号</b> " + " ".join(sig_lines))

    closes = df["close"].astype(float).tolist()
    payload = {
        "symbol": symbol, "period": period,
        "categories": [d.strftime("%Y-%m-%d") for d in df["date"]],
        "kline": [
            [round(float(r.open), 4), round(float(r.close), 4),
             round(float(r.low), 4), round(float(r.high), 4)]
            for r in df.itertuples()
        ],
        "holdEquity": [round(c / closes[0], 4) for c in closes],
        "cards": cards,
        "groups": groups,
        "strategies": strategies,
    }
    for col in ("ma5", "ma10", "ma20"):
        payload[col] = [None if pd.isna(v) else round(float(v), 4) for v in df[col]]

    out_dir = HERE / "reports"
    out_dir.mkdir(parents=True, exist_ok=True)
    out = out_dir / f"chart_{symbol}_{period}.html"
    # echarts 内嵌进 HTML：单文件零外部引用，file:// 双击、IDE 预览、
    # http 服务任何打开方式都稳健（相对引用在部分预览环境下会 404）
    echarts_js = (HERE / "assets" / "echarts.min.js").read_text(encoding="utf-8")
    html = (
        HTML_TEMPLATE.replace("__ECHARTS__", echarts_js)
        .replace("__DATA__", json.dumps(payload, ensure_ascii=False))
        .replace("{title}", f"{symbol} 策略回测报告")
    )
    out.write_text(html, encoding="utf-8")
    print(
        f"[chart] {out.relative_to(HERE)}（{len(df)} 根K线，"
        f"{len(strategies)} 个策略，{len(groups)} 组）"
    )
    return out


def signal_bulls_names(df: pd.DataFrame, bulls: dict) -> list:
    """当前状态卡的信号清单。"""
    return [(None, name) for name in bulls]


def main() -> None:
    parser = argparse.ArgumentParser(description="生成交互式回测报告 HTML")
    parser.add_argument("--symbol", default="562500", help="证券代码")
    parser.add_argument("--period", default="daily", help="daily/weekly/1/5/...")
    args = parser.parse_args()
    make_chart(args.symbol, args.period)


if __name__ == "__main__":
    main()
