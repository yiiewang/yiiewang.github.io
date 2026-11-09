"""随机对照实验：信号策略的表现是 edge，还是"在场时间少"的机械结果？

问题来源：event.py 的差值检验显示信号在收益维度没有区分度（|t|<1.1）。那么
GUIDE 主张的"回撤优于持有 14~18/18"就有两种可能：
  (a) 信号确实能识别风险，有 edge；
  (b) 任何频繁离场的规则都会压回撤——空仓就没有回撤——与预测力无关。

方法：圆形移位（circular shift）。把信号序列整体循环移位 k 天——**在场天数、
在场比例、持仓区间的长度与分布完全不变**，只破坏信号与市场事件的时序对齐。
这是最严格的对照：它控制住了"离场多少"这个混淆因素，只问"离场的时机对不对"。

判定：
  p_dd  = P(随机回撤 >= 真实回撤)   小 → 回撤改善显著优于随机，信号有 edge
  p_ret = P(随机收益 >= 真实收益)   小 → 收益显著优于随机

用法：
  .venv/bin/python perm.py                    # 双确认信号，500 次置换
  .venv/bin/python perm.py --signal MA10x20 --perm 1000
"""

from __future__ import annotations

import argparse

import numpy as np
import pandas as pd

from analyze import buy_and_hold, load, signal_bulls, state_backtest
from scan import UNIVERSE_PATH, load_universe, valid_mask

BENCH = "买入持有"


# ── 信号与回测 ──────────────────────────────────────────────
def make_hold(df: pd.DataFrame, signal: str) -> pd.Series | None:
    """T 日收盘判定的持仓状态（GUIDE 口径，执行由回测引擎 shift(1) 完成）。"""
    bulls = signal_bulls(df)
    if signal == "双确认":
        if "MA5x20" not in bulls or "MACD" not in bulls:
            return None
        return bulls["MA5x20"] & bulls["MACD"]
    if signal in bulls:
        return bulls[signal]
    return None


def fast_backtest(df: pd.DataFrame, hold: pd.Series) -> dict:
    """向量化回测，与 analyze.state_backtest 数学等价（日收益连乘）。

    买入日 close/open、持有日 close/昨收、卖出日 open/昨收、空仓 1.0。
    """
    should = hold.shift(1).fillna(False).astype(bool)
    prev = should.shift(1).fillna(False).astype(bool)
    o = df["open"].astype(float)
    c = df["close"].astype(float)

    day = pd.Series(1.0, index=df.index)
    day[should & ~prev] = c / o
    day[~should & prev] = o / c.shift(1)
    day[should & prev] = c / c.shift(1)

    equity = day.cumprod()
    peak = equity.cummax()
    return {
        "ret": float(equity.iloc[-1] - 1.0),
        "dd": float((equity / peak - 1.0).min()),
        "exposure": float(should.mean()),
        "trades": int((should.astype(int).diff().fillna(0) != 0).sum()),
    }


def verify_equivalence(df: pd.DataFrame, hold: pd.Series) -> None:
    """与 state_backtest 对账（口径一致性自检）。"""
    a = state_backtest(df, hold)
    b = fast_backtest(df, hold)
    print(f"  对账 state_backtest: ret={a['total_return']:+.4f} dd={a['max_dd']:.4f}")
    print(f"  对账 fast_backtest : ret={b['ret']:+.4f} dd={b['dd']:.4f}")
    gap = abs(a["total_return"] - b["ret"]) + abs(a["max_dd"] - b["dd"])
    print(f"  差异 {gap:.2e} -> {'一致' if gap < 1e-9 else '不一致，请勿使用'}")


# ── 置换检验 ────────────────────────────────────────────────
def perm_test(df: pd.DataFrame, hold: pd.Series, n_perm: int,
              seed: int) -> dict:
    n = len(hold)
    real = fast_backtest(df, hold)
    rng = np.random.default_rng(seed)
    ks = rng.choice(np.arange(1, n), size=min(n_perm, n - 1), replace=False)

    vals = np.array(hold.to_numpy(), dtype=bool)
    rets, dds, exps = [], [], []
    for k in ks:
        shifted = pd.Series(np.roll(vals, k), index=hold.index)
        m = fast_backtest(df, shifted)
        rets.append(m["ret"])
        dds.append(m["dd"])
        exps.append(m["exposure"])

    rets = np.asarray(rets)
    dds = np.asarray(dds)
    return {
        "ret": real["ret"],
        "dd": real["dd"],
        "exposure": real["exposure"],
        "trades": real["trades"],
        "rand_ret_med": float(np.median(rets)),
        "rand_dd_med": float(np.median(dds)),
        "p_ret": float((rets >= real["ret"]).mean()),
        "p_dd": float((dds >= real["dd"]).mean()),
        "n_perm": len(ks),
        "exp_diff": float(np.mean(np.abs(np.asarray(exps) - real["exposure"]))),
    }


def main() -> None:
    ap = argparse.ArgumentParser(description="信号策略的随机对照实验")
    ap.add_argument("--universe", default=str(UNIVERSE_PATH))
    ap.add_argument("--signal", default="双确认",
                    choices=["双确认", "MA5x10", "MA5x20", "MA10x20", "MACD"])
    ap.add_argument("--perm", type=int, default=500, help="置换次数")
    ap.add_argument("--seed", type=int, default=20261023)
    ap.add_argument("--verify", action="store_true", help="与 state_backtest 对账后退出")
    args = ap.parse_args()

    cfg = load_universe(UNIVERSE_PATH if args.universe == str(UNIVERSE_PATH) else args.universe)
    period = cfg.get("period", "daily")
    symbols = [s for s in cfg["symbols"] if s["group"] == "equity"]

    if args.verify:
        df = load(symbols[0]["code"], period)
        hold = make_hold(df, args.signal)
        verify_equivalence(df, hold)
        return

    print(f"信号={args.signal}  置换={args.perm}  seed={args.seed}  标的池={len(symbols)}")
    print(f"（对照方法：圆形移位，在场比例严格不变，只打乱时机）\n")

    rows = []
    for s in symbols:
        df = load(s["code"], period)
        vm = valid_mask(df)
        if not vm.any():
            continue
        df = df.loc[df.index[vm][0]:].reset_index(drop=True)
        hold = make_hold(df, args.signal)
        if hold is None:
            continue
        r = perm_test(df, hold, args.perm, args.seed)
        bh_ret, bh_dd = buy_and_hold(df)
        r.update(code=s["code"], name=s.get("name", ""), bh_ret=bh_ret, bh_dd=bh_dd)
        rows.append(r)
    out = pd.DataFrame(rows)

    print(f"{'代码':<10}{'名称':<16}{'在场%':>7}{'策略收益':>10}{'随机中位':>10}{'p_ret':>8}"
          f"{'策略回撤':>10}{'随机中位':>10}{'p_dd':>8}{'持有收益':>10}{'持有回撤':>10}")
    print("-" * 99)
    for _, r in out.iterrows():
        print(f"{r['code']:<10}{str(r['name'])[:14]:<16}{r['exposure']:>7.1%}"
              f"{r['ret']:>+10.1%}{r['rand_ret_med']:>+10.1%}{r['p_ret']:>8.3f}"
              f"{r['dd']:>10.1%}{r['rand_dd_med']:>10.1%}{r['p_dd']:>8.3f}"
              f"{r['bh_ret']:>+10.1%}{r['bh_dd']:>10.1%}")

    print(f"\n{'汇总':<26}"
          f"{'':>7}{out['ret'].mean():>+10.1%}{out['rand_ret_med'].mean():>+10.1%}"
          f"{out['p_ret'].mean():>8.3f}{out['dd'].mean():>10.1%}"
          f"{out['rand_dd_med'].mean():>10.1%}{out['p_dd'].mean():>8.3f}"
          f"{out['bh_ret'].mean():>+10.1%}{out['bh_dd'].mean():>10.1%}")

    print(f"\np_dd < 0.05（回撤显著优于随机）: {(out['p_dd'] < 0.05).sum()} / {len(out)}")
    print(f"p_dd > 0.95（回撤显著差于随机）: {(out['p_dd'] > 0.95).sum()} / {len(out)}")
    print(f"p_ret < 0.05（收益显著优于随机）: {(out['p_ret'] < 0.05).sum()} / {len(out)}")
    print(f"在场比例与随机的最大偏差: {out['exp_diff'].max():.2e}（应≈0，证明对照公平）")

    print("\np_dd 小 = 回撤改善确实是信号的功劳，不是离场的机械结果")
    print("p_dd 大 = 信号挑的离场时机还不如随机离场")


if __name__ == "__main__":
    main()
