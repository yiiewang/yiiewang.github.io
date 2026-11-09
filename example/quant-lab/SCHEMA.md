# 报告数据结构规范（Report Schema）

`chart.py` 产出的报告数据契约。**数据层与展示层通过这个契约解耦**：

```
数据层  make_chart()      → 只算数据，不碰 HTML
        reports/data/{symbol}_{period}.json   ← 独立落盘，任何消费方可直接读
模板层  templates/report.html  → 只依赖下面的字段与三个占位符
渲染层  render()          → 数据 + 模板 → 单文件 HTML
```

换外观只需替换 `templates/report.html`，Python 侧无需改动。

## 版本

`schemaVersion: 1`。结构发生**不兼容**变更时递增（增删顶层字段、改字段含义、
改数组元素顺序）。仅新增可选字段不递增。

模板文件里的占位符只有三个：`__DATA__`（本结构的 JSON）、`__TITLE__`、
`__ECHARTS__`（内联的 echarts.min.js，保证单文件零外部引用）。

## 顶层字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `schemaVersion` | int | 契约版本，当前 1 |
| `symbol` / `period` | str | 证券代码 / 周期（daily / weekly / 1 / 5…） |
| `symbolLabel` / `periodName` | str | 展示用名称（"562500 机器人ETF" / "日线"） |
| `span` | str | 数据区间 "2021-12-29 ~ 2026-10-26" |
| `categories` | string[] | 交易日数组，全部序列的下标基准 |
| `kline` | number[4][] | **每项 `[open, close, low, high]`**——注意不是通用 OHLC 顺序 |
| `ma5` / `ma10` / `ma20` | (number\|null)[] | 均线，预热期为 null |
| `holdEquity` | number[] | 买入持有净值（首日 = 1） |
| `holdMetrics` | Metrics | 买入持有绩效 |
| `cards` | string[] | 顶部状态卡（**含 HTML 片段**，见下方“已知耦合”） |
| `groups` | Group[] | 策略分组，供下拉框使用 |
| `strategies` | {name: Strategy} | 全部策略，键为策略名 |
| `ind` | Indicators | 指标子图数据 |
| `density` | Density | 信号密度（最近 N 天） |
| `state` | State | 当前状态（最新一根 K 线的完整状态） |

### Metrics

`{ret, annual, dd, sharpe, calmar, trades, win}`
收益类单位为**百分数**（`ret: 20.5` 表示 +20.5%）；`sharpe` / `calmar` 可能为 null（零波动时）。

### Group

`{label: str, names: string[]}`——`names` 对应 `strategies` 的键。

### Strategy

| 字段 | 类型 | 说明 |
|---|---|---|
| `kind` | str | `mask`（掩码型）/ `ladder`（三档）/ `switch`（切换系统） |
| `equity` | number[] | 净值曲线 |
| `metrics` | Metrics | 绩效 |
| `buys` / `sells` | [string, number][] | 买卖点 `[日期, 执行价]`，次日开盘成交 |
| `tradeList` | Trade[] | 逐笔明细 |
| `holdCosts` | (number\|null)[] | 每日持仓成本，空仓为 null |
| `desc` | str | 摘要（**含 HTML 片段**） |
| `quality` | str | 信号质量行（**含 HTML 片段**，无数据时为空串） |

`ladder` 另有 `rounds`（按“轮”展示），`switch` 另有 `events`。

**Trade**：`{bi, si, bd, sd, bp, sp, ret, runup, dd, days}`
下标 / 日期 / 买卖价 / 收益% / 最大浮盈% / 最深浮亏% / 持有天数。

### Indicators

`{dif, dea, hist, bull, bear, adx, pdi, mdi}`
- `hist` = dif − dea
- `bull` / `bear` 为背离点 `[[下标, dif值]]`
- 任一列缺失时前端自动隐藏对应子图

### Density

`{buys, sells, warns, rating, nDays, triggers}`
- `rating`：`无信号` / `零星` / `中等密集` / `高度密集`
- `triggers`：`{date, dir, name}`，`dir` ∈ `买入` / `卖出` / `预警`

### State

`{date, close, signals, combos, filters, env, divs, chips}`，后六者均为 **Row 数组**。

**Row** = `{k, v, tone}`
- `k`：标签（如 `"双确认 MA5x20+MACD"`）
- `v`：**已格式化的展示字符串**（如 `"空头 1/2 已42根"`）
- `tone`：`bull` / `bear` / `warn` / `""`，供模板着色

| 组 | 含义 | 示例（k → v） |
|---|---|---|
| `signals` | 各信号源当前多空 | `MA5x10` → `MA5(0.8544) > MA10(0.8318) 多头 已2根` |
| `combos` | 组合信号（策略入场条件） | `双确认 MA5x20+MACD` → `空头 1/2 已42根` |
| `filters` | 过滤器准入 | `ADX>=20` → `通过 ADX=39.6` |
| `env` | 环境与仓位 | `regime` → `趋势↓（1/3） +DI(21.9) vs -DI(25.3)` |
| `divs` | 背离 | `顶背离` → `2026-08-17（44 根前）` |
| `chips` | 状态条摘要 | `回撤` → `-27.2%` |

**注意**：`state` 是**渲染就绪**格式——数值都以字符串呈现（回撤 `-27.2%` 是字符串，
不是 -27.2）。若要拿数值做进一步分析，不要解析这里的字符串，改用
`analyze.collect_current_state()` 的原始结构，它保留了 `bull` / `fast` / `slow` /
`run` / `members_bull` 等字段。二者**同源**（终端报告 `[当前状态]` 段也用它），
判定逻辑只有一份实现。

## 已知耦合（待后续版本清理）

以下字段内嵌 HTML 片段，换模板时若改动展示方式需要同步改 Python：

- `cards[]`
- `strategies[*].desc`
- `strategies[*].quality`

清理方向：改为结构化数据（如 `{label, value, tone}`），由模板负责渲染 HTML。

另有一处命名不一致：`state.signals` 内部用 `fast_label` / `slow_label`（snake_case），
而顶层与策略字段多为 camelCase。后续统一。

## 消费示例

```python
import json
from pathlib import Path

d = json.loads(Path("reports/data/562500_daily.json").read_text(encoding="utf-8"))

# 当前状态：双确认的组合结果（v 为展示字符串）
dual = next(r for r in d["state"]["combos"] if r["k"].startswith("双确认"))
print(d["state"]["date"], dual["k"], "→", dual["v"])

# 各策略收益排名
rank = sorted(d["strategies"].items(), key=lambda kv: -kv[1]["metrics"]["ret"])
for name, s in rank[:5]:
    print(f"{name:<24} {s['metrics']['ret']:+7.1f}%  回撤 {s['metrics']['dd']:.1f}%")

# 最近信号
for t in d["density"]["triggers"][:5]:
    print(t["date"], t["dir"], t["name"])
```

```js
// 浏览器内（模板中 RAW 即本结构）
const st = RAW.state;
console.log(st.date, st.signals.map(s => `${s.name}:${s.bull ? "多" : "空"}`).join(" "));
```

## 文件位置

- 数据：`reports/data/{symbol}_{period}.json`（`reports/` 整体 gitignored，属可再生产物）
- 模板：`templates/report.html`（**入版本库**，因为它不是产物而是源码）
