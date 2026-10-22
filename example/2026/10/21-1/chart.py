"""交互式 HTML K 线图：下拉选择策略，只显示该策略的买卖信号点。

用法（在本目录下）：

    .venv/bin/python chart.py --symbol 000001 --period daily
    .venv/bin/python chart.py --symbol 562500 --period daily

输出 chart_{symbol}_{period}.html（数据内嵌，配合 assets/echarts.min.js
离线打开）。信号点标注在执行价上：▲ 买入 = 策略转多的次日开盘，
▼ 卖出 = 转空的次日开盘（与回测同一无未来函数语义）。切换策略时
同步显示该策略的回测摘要（笔数/胜率/收益/回撤）。

策略清单：单指标（三对均线 + MACD）、双确认（首尾均线 x MACD）、
双确认 + ADX>=20 过滤——全部为"持仓掩码型"策略（三档止盈/网格的
分批出场点结构不同，暂不纳入）。
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
  body { margin: 0; padding: 16px; font: 14px/1.6 -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif; background: #fff; }
  .bar { display: flex; gap: 16px; align-items: center; flex-wrap: wrap; margin-bottom: 10px; }
  select { font-size: 15px; padding: 4px 8px; }
  #summary { color: #333; }
  #summary b { color: #c23531; }
  #chart { width: 100%; height: 640px; }
  .tip { color: #888; font-size: 12px; }
</style>
</head>
<body>
<div class="bar">
  <label>策略：<select id="strategy"></select></label>
  <span id="summary"></span>
</div>
<div id="chart"></div>
<p class="tip">▲ 买入 = 策略转多的次日开盘执行价；▼ 卖出 = 转空的次日开盘执行价（与回测同语义，无未来函数）。滚轮/拖动缩放，悬浮查看明细。</p>
<script src="assets/echarts.min.js"></script>
<script>
const RAW = __DATA__;

const chart = echarts.init(document.getElementById('chart'));
const zoomStart = Math.max(0, 100 * (1 - 250 / RAW.categories.length));

const baseSeries = [
  { name: 'K线', type: 'candlestick', data: RAW.kline,
     itemStyle: { color: '#ef232a', color0: '#14b143',
                  borderColor: '#ef232a', borderColor0: '#14b143' } },
  { name: 'MA5', type: 'line', data: RAW.ma5, showSymbol: false,
     lineStyle: { width: 1, opacity: 0.8 } },
  { name: 'MA10', type: 'line', data: RAW.ma10, showSymbol: false,
     lineStyle: { width: 1, opacity: 0.8 } },
  { name: 'MA20', type: 'line', data: RAW.ma20, showSymbol: false,
     lineStyle: { width: 1.2, opacity: 0.9 } },
  { id: 'buy', name: '买入', type: 'scatter', data: [],
     symbol: 'triangle', symbolSize: 11, z: 5,
     itemStyle: { color: '#e6a23c', borderColor: '#fff', borderWidth: 0.5 } },
  { id: 'sell', name: '卖出', type: 'scatter', data: [],
     symbol: 'triangle', symbolRotate: 180, symbolSize: 11, z: 5,
     itemStyle: { color: '#6c5ce7', borderColor: '#fff', borderWidth: 0.5 } },
];

chart.setOption({
  animation: false,
  tooltip: { trigger: 'axis', axisPointer: { type: 'cross' } },
  legend: { top: 4, data: ['K线', 'MA5', 'MA10', 'MA20', '买入', '卖出'] },
  grid: [{ left: 64, right: 20, top: 36, height: '78%' }],
  xAxis: { type: 'category', data: RAW.categories,
           axisLabel: { formatter: v => v.slice(2) } },
  yAxis: { scale: true, splitArea: { show: true } },
  dataZoom: [
    { type: 'inside', start: zoomStart, end: 100 },
    { type: 'slider', start: zoomStart, end: 100, top: '88%' },
  ],
  series: baseSeries,
});

const select = document.getElementById('strategy');
const summary = document.getElementById('summary');
const names = Object.keys(RAW.strategies);
names.forEach(n => {
  const opt = document.createElement('option');
  opt.value = n; opt.textContent = n;
  select.appendChild(opt);
});

function applyStrategy(name) {
  const s = RAW.strategies[name];
  chart.setOption({
    series: [
      { id: 'buy', data: s.buys },
      { id: 'sell', data: s.sells },
    ],
  });
  summary.innerHTML =
    '<b>' + name + '</b>' +
    '　' + s.trades + ' 笔' +
    '　胜率 ' + s.win_rate +
    '　收益 <b>' + s.total_return + '</b>' +
    '　回撤 ' + s.max_dd;
}
select.addEventListener('change', () => applyStrategy(select.value));
applyStrategy(names.find(n => n.includes('双确认')) || names[0]);
window.addEventListener('resize', () => chart.resize());
</script>
</body>
</html>
"""


def build_strategies(df: pd.DataFrame) -> dict[str, pd.Series]:
    """策略清单：单指标 + 双确认 + 双确认的 ADX 过滤版。"""
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


def mask_to_points(df: pd.DataFrame, hold: pd.Series) -> tuple[list, list]:
    """持仓掩码 -> (买入点, 卖出点)，均标注在翻转次日的开盘执行价。"""
    should = hold.shift(1).fillna(False).astype(bool)
    enter = should & ~should.shift(1).fillna(False).astype(bool)
    exit_ = ~should & should.shift(1).fillna(False).astype(bool)
    opens = df["open"].astype(float)
    dates = [d.strftime("%Y-%m-%d") for d in df["date"]]
    buys = [
        [dates[i], round(float(opens.iloc[i]), 4)]
        for i in df.index[enter]
    ]
    sells = [
        [dates[i], round(float(opens.iloc[i]), 4)]
        for i in df.index[exit_]
    ]
    return buys, sells


def fmt_pct_like(v: float) -> str:
    return f"{v:+.1%}" if v >= 0 else f"{v:.1%}"


def make_chart(symbol: str, period: str) -> Path:
    df = load(symbol, period)
    strategies = build_strategies(df)

    payload: dict = {
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
        "strategies": {},
    }
    for col in ("ma5", "ma10", "ma20"):
        payload[col] = [
            None if pd.isna(v) else round(float(v), 4) for v in df[col]
        ]

    for name, hold in strategies.items():
        buys, sells = mask_to_points(df, hold)
        r = state_backtest(df, hold)
        wr = f"{r['win_rate']:.1%}" if r["win_rate"] is not None else "--"
        payload["strategies"][name] = {
            "buys": buys,
            "sells": sells,
            "trades": r["trades"],
            "win_rate": wr,
            "total_return": fmt_pct_like(r["total_return"]),
            "max_dd": f"{r['max_dd']:.1%}",
        }

    out = HERE / f"chart_{symbol}_{period}.html"
    html = HTML_TEMPLATE.replace("__DATA__", json.dumps(payload, ensure_ascii=False))
    out.write_text(html, encoding="utf-8")
    n_buys = sum(len(v["buys"]) for v in payload["strategies"].values())
    print(
        f"[chart] {out.name}（{len(df)} 根K线，"
        f"{len(strategies)} 个策略共 {n_buys} 个买入点）"
    )
    return out


def main() -> None:
    parser = argparse.ArgumentParser(description="生成交互式策略信号 K 线图")
    parser.add_argument("--symbol", default="562500", help="证券代码")
    parser.add_argument("--period", default="daily", help="daily/weekly/1/5/...")
    args = parser.parse_args()
    make_chart(args.symbol, args.period)


if __name__ == "__main__":
    main()
