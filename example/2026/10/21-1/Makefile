.PHONY: all fetch analyze chart clean help
.DEFAULT_GOAL := help

# ── 可配置参数（命令行覆盖） ──────────────────────────────
SYMBOL   ?= 562500
PERIOD   ?= daily
START    ?= 20200101
END      ?= $(shell date +%Y%m%d)
MA       ?= 5,10,20
MACD     ?= 12,26,9
DENSITY  ?= 5
LIMIT    ?= 10
FORCE    ?= 0
VENV     ?= .venv/bin/python
PY       ?= $(shell [ -x .venv/bin/python ] && echo .venv/bin/python || echo python3)

# FORCE=1 时给 fetch_data.py 追加 --force（强制重拉，忽略缓存）
FORCE_FLAG = $(if $(filter 1,$(FORCE)),--force,)

# ── 一键全流程 ────────────────────────────────────────────
all: fetch analyze chart

# ── 拉取数据（含指标计算，写入 data/{symbol}_{period}.csv） ──
fetch:
	$(PY) fetch_data.py --symbol $(SYMBOL) --period $(PERIOD) \
		--start $(START) --end $(END) --ma $(MA) --macd $(MACD) $(FORCE_FLAG)

# ── 强制重拉（忽略缓存，等价于 FORCE=1） ──────────────────────
refetch:
	$(PY) fetch_data.py --symbol $(SYMBOL) --period $(PERIOD) \
		--start $(START) --end $(END) --ma $(MA) --macd $(MACD) --force

# ── 仅重算指标（不联网，换参数后用） ────────────────────────
recompute:
	$(PY) fetch_data.py --symbol $(SYMBOL) --period $(PERIOD) \
		--ma $(MA) --macd $(MACD) --recompute

# ── 文本分析报告 ───────────────────────────────────────────
analyze:
	$(PY) analyze.py --symbol $(SYMBOL) --period $(PERIOD) \
		--limit $(LIMIT) --density $(DENSITY)

# ── 交互式 HTML 报告 ───────────────────────────────────────
chart:
	$(PY) chart.py --symbol $(SYMBOL) --period $(PERIOD)

# ── 清理产物 ───────────────────────────────────────────────
clean:
	rm -f data/$(SYMBOL)_$(PERIOD).csv
	rm -f reports/chart_$(SYMBOL)_$(PERIOD).html

# ── 帮助 ───────────────────────────────────────────────────
help:
	@echo "用法: make <target> [参数=值 ...]"
	@echo ""
	@echo "一键全流程:"
	@echo "  make all                  拉数据 → 分析 → 生成报告"
	@echo "  make all SYMBOL=002050    换标的"
	@echo ""
	@echo "单步骤:"
	@echo "  make fetch                拉取数据（已缓存则跳过）"
	@echo "  make refetch              强制重拉（--force）"
	@echo "  make recompute            仅重算指标（不联网）"
	@echo "  make analyze              文本分析报告"
	@echo "  make chart                交互式 HTML 报告"
	@echo "  make clean                清理数据与报告"
	@echo ""
	@echo "可配置参数（默认值）:"
	@echo "  SYMBOL=$(SYMBOL)          证券代码"
	@echo "  PERIOD=$(PERIOD)          K线周期（daily/weekly/1/5/15/30/60）"
	@echo "  START=$(START)            起始日期 YYYYMMDD"
	@echo "  END=$(END)                结束日期 YYYYMMDD"
	@echo "  MA=$(MA)                  均线参数"
	@echo "  MACD=$(MACD)              MACD 参数"
	@echo "  DENSITY=$(DENSITY)        信号密度分析天数"
	@echo "  LIMIT=$(LIMIT)            最近信号显示条数"
	@echo "  FORCE=0                   1=强制重拉数据（等价于 refetch）"
	@echo ""
	@echo "示例:"
	@echo ""
	@echo "  # 562500 日线全套流程（默认参数）"
	@echo "  make all"
	@echo ""
	@echo "  # 强制重拉数据再分析（等价于 make refetch && make analyze && make chart）"
	@echo "  make all FORCE=1"
	@echo ""
	@echo "  # 换标的"
	@echo "  make all SYMBOL=002050"
	@echo "  make all SYMBOL=510300 PERIOD=weekly FORCE=1   # 换标的+强制重拉"
	@echo ""
	@echo "  # 换指标参数后重算（不联网）"
	@echo "  make recompute MA=3,10,20 MACD=5,34,7"
	@echo ""
	@echo "  # 换指标参数后强制重拉"
	@echo "  make refetch MA=3,8,21 MACD=5,34,7"
	@echo ""
	@echo "  # 缩短密度窗口看近期变盘"
	@echo "  make analyze DENSITY=3"
	@echo "  make analyze SYMBOL=002050 DENSITY=3 LIMIT=20"
	@echo ""
	@echo "  # 指定起止日期"
	@echo "  make all SYMBOL=562500 START=20220101 END=20231231"
	@echo ""
	@echo "  # 分钟线（东财仅保留近几日）"
	@echo "  make all SYMBOL=562500 PERIOD=5"
	@echo "  make all SYMBOL=002050 PERIOD=15"
	@echo ""
	@echo "  # 组合：换标的 + 换周期 + 换参数 + 短密度"
	@echo "  make all SYMBOL=002050 PERIOD=daily MA=3,10,20 MACD=5,34,7 DENSITY=3"
	@echo ""
	@echo "  # 先拉数据再只看分析（不生成 HTML）"
	@echo "  make fetch && make analyze"
	@echo ""
	@echo "  # 清理后重来"
	@echo "  make clean && make all"
