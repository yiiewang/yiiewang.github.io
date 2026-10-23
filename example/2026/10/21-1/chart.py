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
  #symTitle { font-size: 17px; font-weight: 700; color: #222; margin-bottom: 8px; }
  #symTitle .code { color: #c23531; }
  #symTitle .meta { font-size: 12.5px; font-weight: 400; color: #888; margin-left: 10px; }
  #densityBar { margin-bottom: 8px; font-size: 13px; padding: 5px 12px; border-radius: 3px; }
  #densityBar.density-high { background: #fff0f0; border: 1px solid #f5c6c6; color: #c23531; font-weight: 600; }
  #densityBar.density-mid { background: #fff8e6; border: 1px solid #f0e0b0; color: #8a6d3b; }
  #densityBar.density-low { background: #f5f5f5; border: 1px solid #e5e5e5; color: #888; }
  #densityBar .dt { display: inline-block; margin-right: 10px; font-size: 12px; color: #666; }
  #densityBar .dir-b { color: #e6a23c; }
  #densityBar .dir-s { color: #6c5ce7; }
  #densityBar .dir-w { color: #aaa; }
  .bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; margin-bottom: 8px; }
  select { font-size: 14px; padding: 3px 6px; max-width: 340px; }
  #summary { color: #333; font-size: 13px; }
  #summary b { color: #c23531; }
  .main { display: grid; grid-template-columns: minmax(0, 1fr) 330px; gap: 10px; }
  #mainChart { width: 100%; height: 500px; background: #fff; border: 1px solid #eee; }
  #indChart { width: 100%; height: 250px; background: #fff; border: 1px solid #eee; }
  .subplot-bar { display: flex; gap: 14px; align-items: center; margin: 10px 0 6px;
                 font-size: 12px; color: #666; }
  .subplot-bar label { cursor: pointer; user-select: none; }
  .side { display: flex; flex-direction: column; gap: 6px; min-width: 0;
          height: 0; min-height: 100%; }
  #dayInfo { height: 20px; line-height: 20px; color: #666; font-size: 12.5px;
             white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  #quality { background: #eef6ff; border: 1px solid #c5dcf5; padding: 6px 10px; font-size: 12px; color: #345; }
  #detail { background: #fff8e6; border: 1px solid #f0e0b0; padding: 6px 10px; font-size: 12.5px;
            height: 112px; overflow-y: auto; word-break: break-all; }
  .cmp-item { display: inline-block; white-space: nowrap; margin-right: 8px; }
  #trades { background: #fff; border: 1px solid #eee; flex: 1; min-height: 0; overflow-y: auto; overflow-x: hidden; }
  .t-head, .t-row { display: grid; grid-template-columns: 30px minmax(0, 1fr) 96px 58px 52px; gap: 4px; padding: 3px 8px; font-size: 12px; }
  .t-head > span, .t-row > span { min-width: 0; word-break: break-all; }
  .t-head { position: sticky; top: 0; background: #f0f0f0; font-weight: 600; z-index: 2; }
  .t-row { border-top: 1px solid #f5f5f5; cursor: pointer; }
  .t-row:hover { background: #fdf3e0; }
  .t-row.active { background: #ffe9b8; }
  .tag { font-size: 10px; padding: 0 3px; border-radius: 2px; color: #fff; margin-right: 3px; }
  .tag.t { background: #c23531; } .tag.g { background: #2e8b57; }
  #eqChart { width: 100%; height: 240px; background: #fff; border: 1px solid #eee; margin-top: 6px; }
  #rank { margin-top: 10px; background: #fff; border: 1px solid #e5e5e5; max-height: 420px; overflow-y: auto; }
  #rank table { border-collapse: collapse; width: 100%; font-size: 12.5px; }
  #rank th { position: sticky; top: 0; background: #f0f0f0; font-weight: 600; cursor: pointer;
             padding: 4px 8px; white-space: nowrap; text-align: right; z-index: 2; }
  #rank td { padding: 3px 8px; border-top: 1px solid #f5f5f5; white-space: nowrap; text-align: right; }
  #rank th:nth-child(1), #rank th:nth-child(2),
  #rank td:nth-child(1), #rank td:nth-child(2) { text-align: left; }
  #rank tbody tr { cursor: pointer; }
  #rank tbody tr:hover { background: #fdf3e0; }
  #rank tbody tr.active { background: #ffe9b8; }
  .tip { color: #888; font-size: 12px; margin-top: 5px; }
  .pos { color: #c23531; } .neg { color: #2e8b57; }
</style>
</head>
<body>
<div id="symTitle"></div>
<div class="cards" id="cards"></div>
<div id="densityBar"></div>
<div class="bar">
  <label>策略：<select id="strategy"></select></label>
  <span id="summary"></span>
</div>
<div class="main">
  <div>
    <div id="mainChart"></div>
    <div id="dayInfo" style="margin-top:4px">点击K线任意一天可查看该日持仓与浮盈</div>
    <div class="subplot-bar">浮窗：
      <label><input type="checkbox" id="tipMain" checked>主图</label>
      <label><input type="checkbox" id="tipInd" checked>指标</label>
      <label><input type="checkbox" id="tipEq">净值</label>
    </div>
    <div id="indChart"></div>
    <div id="eqChart"></div>
  </div>
  <div class="side">
    <div id="detail"></div>
    <div id="quality"></div>
    <div id="trades"></div>
  </div>
</div>
<div id="rank"></div>
<p class="tip">▲ 买入 = 转多的次日开盘执行价；▼ 卖出 = 转空的次日开盘（与回测同语义）。点击列表定位 K 线区间并显示同期跨策略对比；切换系统选中时主图背景为模式条带（红=趋势/绿=网格/灰=现金）；底部净值图与主图联动。</p>
<script>__ECHARTS__</script>
<script>
const RAW = __DATA__;

const mainChart = echarts.init(document.getElementById('mainChart'));
const eqChart = echarts.init(document.getElementById('eqChart'));
const indBox = document.getElementById('indChart');
const indChart = RAW.ind && (RAW.ind.dif || RAW.ind.adx) ? echarts.init(indBox) : null;
if (!indChart) {
  indBox.style.display = 'none';
  document.getElementById('tipInd').parentElement.style.display = 'none';
}
echarts.connect([mainChart, eqChart, indChart].filter(Boolean));
const N = RAW.categories.length;
const zoomStart = Math.max(0, 100 * (1 - 250 / N));
const cat = { type: 'category', data: RAW.categories, axisLabel: { formatter: v => v.slice(2) } };

// 证券标题栏：代码 + 名称 + 周期 + 数据区间
document.getElementById('symTitle').innerHTML =
  '<span class="code">' + RAW.symbolLabel + '</span> 策略回测报告'
  + '<span class="meta">' + (RAW.periodName || RAW.period) + '　' + (RAW.span || '')
  + '　' + RAW.categories.length + ' 根K线</span>';

// 顶部状态卡
document.getElementById('cards').innerHTML =
  RAW.cards.map(c => '<div class="card">' + c + '</div>').join('');

// 信号密度状态条
(function() {
  const d = RAW.density;
  const bar = document.getElementById('densityBar');
  const cls = d.buys + d.sells >= 5 ? 'density-high' : d.buys + d.sells >= 3 ? 'density-mid' : 'density-low';
  let html = `信号密度（最近${d.nDays}天）：买入 <b>${d.buys}</b> / 卖出 <b>${d.sells}</b> / 预警 ${d.warns} → <b>${d.rating}</b>`;
  if (d.triggers.length) {
    const dirs = {'买入': 'dir-b', '卖出': 'dir-s', '预警': 'dir-w'};
    html += '<br>' + d.triggers.slice(0, 12).map(t =>
      `<span class="dt">${t.date} <span class="${dirs[t.dir]}">${t.dir}</span> ${t.name}</span>`).join('');
    if (d.triggers.length > 12) html += `<span class="dt">…共 ${d.triggers.length} 条</span>`;
  }
  bar.className = cls;
  bar.innerHTML = html;
})();

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
    { id: 'buy', name: '买入', type: 'scatter', data: [], symbol: 'triangle', symbolSize: 12, z: 5,
      itemStyle: { color: '#e6a23c', borderColor: '#fff', borderWidth: .5 },
      label: { show: true, formatter: 'B', position: 'bottom', color: '#e6a23c', fontWeight: 'bold', fontSize: 12 } },
    { id: 'sell', name: '卖出', type: 'scatter', data: [], symbol: 'triangle', symbolRotate: 180, symbolSize: 12, z: 5,
      itemStyle: { color: '#6c5ce7', borderColor: '#fff', borderWidth: .5 },
      label: { show: true, formatter: 'S', position: 'top', color: '#6c5ce7', fontWeight: 'bold', fontSize: 12 } },
    { id: 'fbuy', type: 'scatter', data: [], symbol: 'triangle', symbolSize: 20, z: 6,
      itemStyle: { color: '#ff8c00', borderColor: '#333', borderWidth: 1 } },
    { id: 'fsell', type: 'scatter', data: [], symbol: 'triangle', symbolRotate: 180, symbolSize: 20, z: 6,
      itemStyle: { color: '#4834d4', borderColor: '#333', borderWidth: 1 } },
    { id: 'density', name: '信号密度', type: 'scatter', data: [],
      symbol: 'pin', symbolSize: 28, z: 7,
      itemStyle: { color: 'rgba(230,162,60,.7)', borderColor: '#c23531', borderWidth: 1 },
      label: { show: true, formatter: p => p.data.value || '', position: 'top',
               color: '#c23531', fontWeight: 'bold', fontSize: 10 } },
  ],
});

// 信号密度标记：最近 N 天有触发的日期在主图打标记
(function() {
  const d = RAW.density;
  if (!d.triggers.length) return;
  // 按日期聚合：同一天多个触发合并，标记数量
  const byDate = {};
  d.triggers.forEach(t => { byDate[t.date] = (byDate[t.date] || 0) + 1; });
  const cats = RAW.categories;
  const data = Object.entries(byDate).map(([date, count]) => {
    const idx = cats.indexOf(date);
    return idx >= 0 ? { value: [idx, count], name: date } : null;
  }).filter(Boolean);
  if (data.length) mainChart.setOption({ series: [{ id: 'density', data }] });
})();

const eqNames = Object.keys(RAW.strategies).concat(['买入持有']);
eqChart.setOption({
  animation: false,
  tooltip: { trigger: 'axis', showContent: false },
  legend: { top: 2, type: 'scroll', show: false },
  grid: { left: 60, right: 16, top: 30, bottom: 24 },
  xAxis: cat,
  yAxis: { type: 'log', scale: true, axisLabel: { formatter: v => v.toFixed(2) } },
  dataZoom: [{ type: 'inside', start: zoomStart, end: 100, zoomOnMouseWheel: 'shift' }],
  series: eqNames.map(n => ({
    name: n, type: 'line', showSymbol: false,
    lineStyle: { width: n.includes('双确认') || n === '买入持有' ? 1.8 : 1, opacity: .85 },
    data: n === '买入持有' ? RAW.holdEquity : RAW.strategies[n].equity,
  })),
});

// 指标子图：MACD（DIF/DEA/柱 + 背离点）与 ADX/DI（20/25 阈值线）
if (indChart) {
  const IND = RAW.ind;
  const grids = [], xAxes = [], yAxes = [], series = [], lg = [];
  let g = -1;
  if (IND.dif) {
    g++;
    grids.push({ left: 60, right: 16, top: 26, height: 95 });
    xAxes.push({ type: 'category', gridIndex: g, data: RAW.categories,
                 axisLabel: { show: false }, axisTick: { show: false } });
    yAxes.push({ gridIndex: g, scale: true });
    lg.push('DIF', 'DEA', 'MACD柱');
    series.push(
      { name: 'DIF', type: 'line', xAxisIndex: g, yAxisIndex: g, data: IND.dif, showSymbol: false,
        itemStyle: { color: '#5470c6' }, lineStyle: { width: 1.3 },
        markLine: { silent: true, symbol: 'none', label: { show: false },
                    lineStyle: { type: 'dashed', color: '#bbb' }, data: [{ yAxis: 0 }] } },
      { name: 'DEA', type: 'line', xAxisIndex: g, yAxisIndex: g, data: IND.dea, showSymbol: false,
        itemStyle: { color: '#e6a23c' }, lineStyle: { width: 1.1 } },
      { name: 'MACD柱', type: 'bar', xAxisIndex: g, yAxisIndex: g, data: IND.hist, barWidth: '60%',
        itemStyle: { color: p => p.value == null || p.value >= 0 ? '#ef232a' : '#14b143' } });
    if (IND.bull) {
      lg.push('底背离');
      series.push({ name: '底背离', type: 'scatter', xAxisIndex: g, yAxisIndex: g, data: IND.bull,
        symbolSize: 6, z: 6, itemStyle: { color: '#14b143', borderColor: '#fff', borderWidth: 1 } });
    }
    if (IND.bear) {
      lg.push('顶背离');
      series.push({ name: '顶背离', type: 'scatter', xAxisIndex: g, yAxisIndex: g, data: IND.bear,
        symbolSize: 6, z: 6, itemStyle: { color: '#ef232a', borderColor: '#fff', borderWidth: 1 } });
    }
  }
  if (IND.adx) {
    g++;
    grids.push({ left: 60, right: 16, top: 134, height: 95 });
    xAxes.push({ type: 'category', gridIndex: g, data: RAW.categories,
                 axisLabel: { formatter: v => v.slice(2) } });
    yAxes.push({ gridIndex: g });
    lg.push('ADX', '+DI', '-DI');
    series.push(
      { name: 'ADX', type: 'line', xAxisIndex: g, yAxisIndex: g, data: IND.adx, showSymbol: false,
        itemStyle: { color: '#333' }, lineStyle: { width: 1.6 },
        markLine: { silent: true, symbol: 'none', label: { show: false },
                    lineStyle: { type: 'dashed', color: '#999' }, data: [{ yAxis: 20 }, { yAxis: 25 }] } },
      { name: '+DI', type: 'line', xAxisIndex: g, yAxisIndex: g, data: IND.pdi, showSymbol: false,
        itemStyle: { color: '#e6a23c' }, lineStyle: { width: 1 } },
      { name: '-DI', type: 'line', xAxisIndex: g, yAxisIndex: g, data: IND.mdi, showSymbol: false,
        itemStyle: { color: '#2e8b57' }, lineStyle: { width: 1 } });
  }
  indChart.setOption({
    animation: false,
    tooltip: { trigger: 'axis', axisPointer: { type: 'cross' } },
    axisPointer: { link: [{ xAxisIndex: 'all' }] },
    legend: { top: 2, type: 'scroll', data: lg },
    grid: grids, xAxis: xAxes, yAxis: yAxes,
    dataZoom: [{ type: 'inside', xAxisIndex: xAxes.map((_, i) => i), start: zoomStart, end: 100, zoomOnMouseWheel: 'shift' }],
    series,
  });
}

// 浮窗开关：控制各图 tooltip 浮层显隐（关闭时保留十字准线定位）
document.getElementById('tipMain').addEventListener('change', e =>
  mainChart.setOption({ tooltip: { showContent: e.target.checked } }));
document.getElementById('tipInd').addEventListener('change', e => {
  if (indChart) indChart.setOption({ tooltip: { showContent: e.target.checked } });
});
document.getElementById('tipEq').addEventListener('change', e =>
  eqChart.setOption({ tooltip: { showContent: e.target.checked } }));

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
      return '<span class="cmp-item ' + cls(r) + '">' + n + ' ' + pct(r) + '</span>'; })
    .join('');
  const hr = (RAW.holdEquity[si] / RAW.holdEquity[bi] - 1) * 100;
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
    { id: 'buy', name: name + ' 买入', data: s.buys || [] },
    { id: 'sell', name: name + ' 卖出', data: s.sells || [] },
    { id: 'fbuy', data: [] }, { id: 'fsell', data: [] },
    { name: 'K线', markArea: { silent: true, data: (s.areas || []).map(a => [
      { xAxis: RAW.categories[a[0]], itemStyle: { color: a[2] } },
      { xAxis: RAW.categories[a[1]] }]) } },
  ]});
  summary.innerHTML = '<b>' + name + '</b>　' + (s.desc || '');
  quality.innerHTML = s.quality || '';
  quality.style.display = s.quality ? 'block' : 'none';
  detail.innerHTML = '<span style="color:#999">点击下方列表中的交易行：定位 K 线区间，并显示同期跨策略收益对比</span>';

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
    markPoints(null, 0, null, 0);
  }
  renderRank();
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

// ===== 策略排名表：全策略指标对比 + 综合分 =====
const rankBox = document.getElementById('rank');
const rankRows = Object.keys(RAW.strategies)
  .map(n => ({ name: n, m: RAW.strategies[n].metrics }))
  .concat([{ name: '买入持有', m: RAW.holdMetrics }]);
// 综合分 = 五项指标各自排名(0 起)的平均，越小越优；缺失值按最差处理
const WORST = { ret: -Infinity, annual: -Infinity, dd: Infinity, sharpe: -Infinity, calmar: -Infinity };
const vOf = (r, k) => r.m[k] == null ? WORST[k] : r.m[k];
[
  ['ret', (a, b) => b - a],
  ['annual', (a, b) => b - a],
  ['dd', (a, b) => Math.abs(a) - Math.abs(b)],
  ['sharpe', (a, b) => b - a],
  ['calmar', (a, b) => b - a],
].forEach(([k, cmp]) => {
  [...rankRows].sort((x, y) => cmp(vOf(x, k), vOf(y, k)))
    .forEach((r, i) => { r.score = (r.score || 0) + i; });
});
rankRows.forEach(r => { r.score /= 5; });

const RANK_COLS = [
  { k: 'name', label: '策略', dir: 1, get: r => r.name, fmt: v => v },
  { k: 'ret', label: '总收益', dir: -1, get: r => r.m.ret,
    fmt: v => '<span class="' + cls(v) + '">' + pct(v) + '</span>' },
  { k: 'annual', label: '年化', dir: -1, get: r => r.m.annual,
    fmt: v => '<span class="' + cls(v) + '">' + pct(v) + '</span>' },
  { k: 'dd', label: '最大回撤', dir: 1, get: r => r.m.dd,
    fmt: v => '<span class="neg">' + v.toFixed(1) + '%</span>' },
  { k: 'sharpe', label: '夏普', dir: -1, get: r => r.m.sharpe, fmt: v => v == null ? '--' : v.toFixed(2) },
  { k: 'calmar', label: '卡尔玛', dir: -1, get: r => r.m.calmar, fmt: v => v == null ? '--' : v.toFixed(2) },
  { k: 'trades', label: '交易数', dir: -1, get: r => r.m.trades, fmt: v => v },
  { k: 'win', label: '胜率', dir: -1, get: r => r.m.win, fmt: v => v == null ? '--' : v.toFixed(1) + '%' },
  { k: 'score', label: '综合分', dir: 1, get: r => r.score, fmt: v => v.toFixed(1) },
];
let rankKey = 'score', rankDir = 1;

function renderRank() {
  const col = RANK_COLS.find(c => c.k === rankKey);
  const rows = [...rankRows].sort((a, b) => {
    const av = col.get(a), bv = col.get(b);
    if (av === null && bv === null) return a.name.localeCompare(b.name, 'zh');
    if (av === null) return 1;
    if (bv === null) return -1;
    const d = col.k === 'name' ? String(av).localeCompare(String(bv), 'zh')
      : col.k === 'dd' ? Math.abs(av) - Math.abs(bv) : av - bv;
    return (d !== 0 ? d : a.name.localeCompare(b.name, 'zh')) * rankDir;
  });
  const th = RANK_COLS.map(c =>
    '<th data-k="' + c.k + '">' + c.label +
    (c.k === rankKey ? (rankDir === 1 ? ' ▲' : ' ▼') : '') + '</th>').join('');
  const trs = rows.map((r, i) =>
    '<tr data-n="' + r.name + '"' + (r.name === current ? ' class="active"' : '') + '>' +
    '<td>' + (i + 1) + '</td>' +
    RANK_COLS.map(c => '<td>' + c.fmt(c.get(r)) + '</td>').join('') +
    '</tr>').join('');
  rankBox.innerHTML = '<table><thead><tr><th>排名</th>' + th + '</tr></thead><tbody>' + trs + '</tbody></table>';
  rankBox.querySelectorAll('th[data-k]').forEach(el => {
    el.onclick = () => {
      const k = el.dataset.k;
      if (rankKey === k) rankDir = -rankDir;
      else { rankKey = k; rankDir = RANK_COLS.find(c => c.k === k).dir; }
      renderRank();
    };
  });
  rankBox.querySelectorAll('tbody tr').forEach(tr => {
    const n = tr.dataset.n;
    if (n in RAW.strategies) tr.onclick = () => applyStrategy(n);
  });
}

select.addEventListener('change', () => applyStrategy(select.value));
applyStrategy('三档+冷却协议');
window.addEventListener('resize', () => { mainChart.resize(); eqChart.resize(); if (indChart) indChart.resize(); });
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
        .replace("{title}", f"{symbol_label(symbol)} 策略回测报告")
    )
    out.write_text(html, encoding="utf-8")
    print(
        f"[chart] {out.relative_to(HERE)}（{len(df)} 根K线，"
        f"{len(strategies)} 个策略，{len(groups)} 组）"
    )
    return out


def main() -> None:
    parser = argparse.ArgumentParser(description="生成交互式回测报告 HTML")
    parser.add_argument("--symbol", default="562500", help="证券代码")
    parser.add_argument("--period", default="daily", help="daily/weekly/1/5/...")
    args = parser.parse_args()
    make_chart(args.symbol, args.period)


if __name__ == "__main__":
    main()
