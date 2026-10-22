"""交互式 HTML 回测报告：K 线信号 + 操作列表 + 跨策略对比。

用法（在本目录下）：

    .venv/bin/python chart.py --symbol 562500 --period daily
    .venv/bin/python chart.py --symbol 000001 --period daily   # 上证指数

输出 chart_{symbol}_{period}.html（配合 assets/echarts.min.js 离线打开）。

交互能力：
- 下拉切换策略，K 线只显示该策略的买卖信号点（▲/▼ 标注执行价）
- 右侧操作列表逐笔展示交易（买卖日期/价格/收益/持有天数/期间浮盈浮亏），
  点击任意一笔 -> K 线自动缩放聚焦到该笔区间并高亮，同时显示
  同一时间窗口内全部策略的区间收益对比
- 点击 K 线任意一天 -> 显示该日该策略的持仓状态与浮动盈亏
- 底部净值曲线对比图（全部策略 + 买入持有），与主图缩放联动
"""

from __future__ import annotations

import argparse
import json
from pathlib import Path

import pandas as pd

from analyze import load, ma_columns, signal_bulls, state_backtest

HERE = Path(__file__).resolve().parent

HTML_TEMPLATE = """<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="utf-8">
<title>{title}</title>
<style>
  * { box-sizing: border-box; }
  body { margin: 0; padding: 14px; font: 14px/1.6 -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif; background: #fafafa; }
  .bar { display: flex; gap: 14px; align-items: center; flex-wrap: wrap; margin-bottom: 10px; }
  select { font-size: 15px; padding: 4px 8px; }
  #summary { color: #333; }
  #summary b { color: #c23531; }
  .main { display: grid; grid-template-columns: 1fr 340px; gap: 12px; }
  #mainChart { width: 100%; height: 520px; background: #fff; border: 1px solid #eee; }
  .side { display: flex; flex-direction: column; gap: 8px; }
  #dayInfo { min-height: 20px; color: #666; font-size: 13px; }
  #detail { background: #fff8e6; border: 1px solid #f0e0b0; padding: 8px 10px; font-size: 13px; display: none; }
  #detail .cmp { color: #555; }
  .pos { color: #c23531; } .neg { color: #2e8b57; }
  #trades { background: #fff; border: 1px solid #eee; height: 434px; overflow-y: auto; }
  .t-head, .t-row { display: grid; grid-template-columns: 34px 1fr 92px 62px 44px; gap: 4px; padding: 4px 8px; font-size: 12.5px; }
  .t-head { position: sticky; top: 0; background: #f0f0f0; font-weight: 600; z-index: 2; }
  .t-row { border-top: 1px solid #f5f5f5; cursor: pointer; }
  .t-row:hover { background: #fdf3e0; }
  .t-row.active { background: #ffe9b8; }
  #eqChart { width: 100%; height: 250px; background: #fff; border: 1px solid #eee; margin-top: 12px; }
  .tip { color: #888; font-size: 12px; margin-top: 6px; }
</style>
</head>
<body>
<div class="bar">
  <label>策略：<select id="strategy"></select></label>
  <span id="summary"></span>
  <span id="dayInfo">点击K线任意一天可查看该日持仓与浮盈</span>
</div>
<div class="main">
  <div id="mainChart"></div>
  <div class="side">
    <div id="detail"></div>
    <div id="trades"></div>
  </div>
</div>
<div id="eqChart"></div>
<p class="tip">▲ 买入 = 策略转多的次日开盘执行价；▼ 卖出 = 转空的次日开盘（与回测同语义）。点击右侧任意一笔交易：K 线聚焦该区间 + 显示同期各策略收益对比。底部净值图与主图缩放联动。</p>
<script src="assets/echarts.min.js"></script>
<script>
const RAW = __DATA__;

const mainChart = echarts.init(document.getElementById('mainChart'));
const eqChart = echarts.init(document.getElementById('eqChart'));
echarts.connect([mainChart, eqChart]);
const N = RAW.categories.length;
const zoomStart = Math.max(0, 100 * (1 - 250 / N));
const cat = { type: 'category', data: RAW.categories, axisLabel: { formatter: v => v.slice(2) } };

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
    { name: 'K线', type: 'candlestick', data: RAW.kline,
      itemStyle: { color: '#ef232a', color0: '#14b143', borderColor: '#ef232a', borderColor0: '#14b143' } },
    { name: 'MA5', type: 'line', data: RAW.ma5, showSymbol: false, lineStyle: { width: 1, opacity: .8 } },
    { name: 'MA10', type: 'line', data: RAW.ma10, showSymbol: false, lineStyle: { width: 1, opacity: .8 } },
    { name: 'MA20', type: 'line', data: RAW.ma20, showSymbol: false, lineStyle: { width: 1.2, opacity: .9 } },
    { id: 'buy', name: '买入', type: 'scatter', data: [], symbol: 'triangle', symbolSize: 10, z: 5,
      itemStyle: { color: '#e6a23c', borderColor: '#fff', borderWidth: .5 } },
    { id: 'sell', name: '卖出', type: 'scatter', data: [], symbol: 'triangle', symbolRotate: 180, symbolSize: 10, z: 5,
      itemStyle: { color: '#6c5ce7', borderColor: '#fff', borderWidth: .5 } },
    { id: 'fbuy', type: 'scatter', data: [], symbol: 'triangle', symbolSize: 20, z: 6,
      itemStyle: { color: '#ff8c00', borderColor: '#333', borderWidth: 1 }, label: { show: true, position: 'bottom', formatter: '买', color: '#ff8c00' } },
    { id: 'fsell', type: 'scatter', data: [], symbol: 'triangle', symbolRotate: 180, symbolSize: 20, z: 6,
      itemStyle: { color: '#4834d4', borderColor: '#333', borderWidth: 1 }, label: { show: true, position: 'top', formatter: '卖', color: '#4834d4' } },
  ],
});

const eqNames = Object.keys(RAW.strategies).concat(['买入持有']);
eqChart.setOption({
  animation: false,
  tooltip: { trigger: 'axis' },
  legend: { top: 2, type: 'scroll' },
  grid: { left: 60, right: 16, top: 32, bottom: 24 },
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
const tradesBox = document.getElementById('trades');
const names = Object.keys(RAW.strategies);
names.forEach(n => { const o = document.createElement('option'); o.value = n; o.textContent = n; select.appendChild(o); });
let current = null;

const pct = v => (v >= 0 ? '+' : '') + v.toFixed(2) + '%';
const cls = v => v >= 0 ? 'pos' : 'neg';

function applyStrategy(name) {
  current = name;
  const s = RAW.strategies[name];
  mainChart.setOption({ series: [
    { id: 'buy', data: s.buys },
    { id: 'sell', data: s.sells },
    { id: 'fbuy', data: [] }, { id: 'fsell', data: [] },
  ]});
  summary.innerHTML = '<b>' + name + '</b>　' + s.trades + ' 笔　胜率 ' + s.win_rate
    + '　收益 <b>' + s.total_return + '</b>　回撤 ' + s.max_dd;
  detail.style.display = 'none';
  dayInfo.textContent = '点击K线任意一天可查看该日持仓与浮盈';
  tradesBox.innerHTML = '<div class="t-head"><span>#</span><span>买入 → 卖出</span><span>价格</span><span>收益</span><span>天数</span></div>';
  s.tradeList.forEach((t, i) => {
    const row = document.createElement('div');
    row.className = 't-row';
    row.innerHTML = '<span>' + (i + 1) + '</span>'
      + '<span>' + t.bd.slice(5) + ' → ' + t.sd.slice(5) + '</span>'
      + '<span>' + t.bp + '→' + t.sp + '</span>'
      + '<span class="' + cls(t.ret) + '">' + pct(t.ret) + '</span>'
      + '<span>' + t.days + 'd</span>';
    row.onclick = () => focusTrade(t, row);
    tradesBox.appendChild(row);
  });
}

function focusTrade(t, row) {
  document.querySelectorAll('.t-row.active').forEach(r => r.classList.remove('active'));
  row.classList.add('active');
  const pad = 30;
  mainChart.dispatchAction({ type: 'dataZoom',
    start: Math.max(0, (t.bi - pad) * 100 / N),
    end: Math.min(100, (t.si + pad) * 100 / N) });
  mainChart.setOption({ series: [
    { id: 'fbuy', data: [[t.bd, t.bp]] },
    { id: 'fsell', data: [[t.sd, t.sp]] },
  ]});
  const cmp = Object.entries(RAW.strategies).map(([n, st]) => {
    const r = (st.equity[t.si] / st.equity[t.bi] - 1) * 100;
    return '<span class="' + cls(r) + '" style="margin-right:8px">' + n + ' ' + pct(r) + '</span>';
  });
  const hr = (RAW.holdEquity[t.si] / RAW.holdEquity[t.bi] - 1) * 100;
  detail.style.display = 'block';
  detail.innerHTML = '<b>本笔 ' + pct(t.ret) + '</b>（' + t.bd + ' ' + t.bp + ' → ' + t.sd + ' ' + t.sp
    + '，' + t.days + ' 天）　期间最高浮盈 ' + pct(t.runup) + '　最深浮亏 ' + pct(t.dd)
    + '<br><span class="cmp">同区间对比：' + cmp.join('')
    + '<span class="' + cls(hr) + '">买入持有 ' + pct(hr) + '</span></span>';
}

mainChart.on('click', params => {
  if (params.seriesType !== 'candlestick') return;
  const i = params.dataIndex;
  const cost = RAW.strategies[current].holdCosts[i];
  const c = params.data[1];
  dayInfo.innerHTML = cost == null
    ? RAW.categories[i] + '：空仓（close=' + c + '）'
    : RAW.categories[i] + '：持仓 成本 ' + cost + '　close=' + c
      + '　浮动盈亏 <b class="' + cls((c / cost - 1) * 100) + '">' + pct((c / cost - 1) * 100) + '</b>';
});

select.addEventListener('change', () => applyStrategy(select.value));
applyStrategy(names.find(n => n.includes('双确认')) || names[0]);
window.addEventListener('resize', () => { mainChart.resize(); eqChart.resize(); });
</script>
</body>
</html>
"""


def build_strategies(df: pd.DataFrame) -> dict[str, pd.Series]:
    """策略清单：单指标 + 双确认 + 双确认的 ADX 过滤版（持仓掩码型）。"""
    bulls = signal_bulls(df)
    strategies: dict[str, pd.Series] = dict(bulls)
    cols = ma_columns(df)
    if len(cols) >= 2 and "MACD" in bulls:
        pair = f"MA{cols[0][2:]}x{cols[-1][2:]}"
        combo = bulls[pair] & bulls["MACD"]
        strategies[f"{pair}+MACD(双确认)"] = combo
        if "adx" in df.columns:
            strategies[f"{pair}+MACD+ADX>=20(过滤)"] = combo & (df["adx"] >= 20)
    return strategies


def build_strategy_payload(df: pd.DataFrame, hold: pd.Series) -> dict:
    """策略完整数据：逐笔交易配对、逐日持仓成本、逐日净值。"""
    should = hold.shift(1).fillna(False).astype(bool)
    opens = df["open"].astype(float).tolist()
    closes = df["close"].astype(float).tolist()
    dates = [d.strftime("%Y-%m-%d") for d in df["date"]]
    n = len(df)
    s = should.tolist()

    buys, sells = [], []
    hold_costs: list[float | None] = [None] * n
    trade_list: list[dict] = []
    entry = None  # (idx, price)

    def close_trade(exit_i: int, exit_p: float) -> None:
        bi, bp = entry
        seg = closes[bi: exit_i + 1]
        trade_list.append(
            {
                "bi": bi, "si": exit_i,
                "bd": dates[bi], "sd": dates[exit_i],
                "bp": round(bp, 4), "sp": round(exit_p, 4),
                "ret": round((exit_p / bp - 1) * 100, 2),
                "runup": round((max(seg) / bp - 1) * 100, 2),
                "dd": round((min(seg) / bp - 1) * 100, 2),
                "days": exit_i - bi,
            }
        )

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
    if entry is not None:  # 期末持仓按收盘平仓
        close_trade(n - 1, closes[-1])

    r = state_backtest(df, hold)
    equity = r["equity_curve"].iloc[1:].tolist()  # 对齐每日收盘
    hold_eq = closes[0]
    return {
        "buys": buys,
        "sells": sells,
        "tradeList": trade_list,
        "holdCosts": hold_costs,
        "equity": [round(v, 4) for v in equity],
        "trades": r["trades"],
        "win_rate": f"{r['win_rate']:.1%}" if r["win_rate"] is not None else "--",
        "total_return": f"{r['total_return']:+.1%}",
        "max_dd": f"{r['max_dd']:.1%}",
    }


def make_chart(symbol: str, period: str) -> Path:
    df = load(symbol, period)
    strategies = build_strategies(df)
    closes = df["close"].astype(float).tolist()

    payload = {
        "symbol": symbol,
        "period": period,
        "categories": [d.strftime("%Y-%m-%d") for d in df["date"]],
        "kline": [
            [
                round(float(r.open), 4),
                round(float(r.close), 4),
                round(float(r.low), 4),
                round(float(r.high), 4),
            ]
            for r in df.itertuples()
        ],
        "holdEquity": [round(c / closes[0], 4) for c in closes],
        "strategies": {
            name: build_strategy_payload(df, hold)
            for name, hold in strategies.items()
        },
    }
    for col in ("ma5", "ma10", "ma20"):
        payload[col] = [
            None if pd.isna(v) else round(float(v), 4) for v in df[col]
        ]

    out = HERE / f"chart_{symbol}_{period}.html"
    html = HTML_TEMPLATE.replace("__DATA__", json.dumps(payload, ensure_ascii=False))
    html = html.replace("{title}", f"{symbol} 策略回测报告")
    out.write_text(html, encoding="utf-8")
    n_trades = sum(v["trades"] for v in payload["strategies"].values())
    print(
        f"[chart] {out.name}（{len(df)} 根K线，{len(strategies)} 个策略，"
        f"共 {n_trades} 笔交易明细）"
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
