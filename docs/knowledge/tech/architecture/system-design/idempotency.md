---
title: "Go 服务幂等设计实战"
created: 2026-07-01
updated: 2026-07-01
domain: tech
tags:
  - Go
  - 后端工程
  - 幂等设计
  - 分布式系统
  - 数据库
difficulty: intermediate
summary: "以支付回调为例，拆解幂等设计的四层防线：业务语义定义、唯一约束 + 事务、失败恢复策略、可观测性监控。"
source:
  - type: blog
    url: /blog/posts/2026/02/28/
---

# Go 服务幂等设计实战

> 幂等不是加一行 `if` 判断，而是对**业务状态机的约束**。哪个字段代表"动作已生效"，哪些操作必须在同一事务完成，重复请求返回什么——这三个问题必须回答。

## 问题场景

典型路径：

```
用户支付 → 第三方回调 → 更新订单 + 扣库存 → 返回成功
```

第三方可能因网络抖动重复回调，同一订单短时间内并发处理多次。症状：

- 库存被重复扣减
- 积分/优惠券被重复发放
- 下游收到多条语义相同的事件

## 四层防重设计

### 第一层：定义幂等键

以**业务动作**而非 HTTP 请求为单位：

| 推荐 | 不推荐 |
|------|--------|
| `biz_type + biz_id`（如 `PAY_SUCCESS:order_123`） | 随机请求 ID（每次不一样） |

```sql
CREATE TABLE idempotency_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    biz_type VARCHAR(64) NOT NULL,
    biz_id VARCHAR(128) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'processing',
    response_body TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_biz (biz_type, biz_id)  -- 第一道硬防线
);
```

### 第二层：唯一约束 + 事务 + 状态机

```go
func (s *Service) HandlePaySuccess(ctx context.Context, orderID string) error {
    // ① 占坑：唯一索引天然防并发
    ok, _ := s.repo.TryCreateIdempotency(ctx, "PAY_SUCCESS", orderID)
    if !ok { return nil } // 已处理过

    // ② 同一事务完成业务更新
    return s.repo.WithTx(ctx, func(tx Tx) error {
        order, _ := tx.GetOrderForUpdate(orderID)
        if order.Status == "paid" { return nil }
        tx.DeductInventory(order.Items)
        tx.MarkOrderPaid(orderID)
        return tx.MarkIdempotencyDone("PAY_SUCCESS", orderID)
    })
}
```

关键点：
- `TryCreateIdempotency` 依赖唯一索引，天然防并发
- 扣库存与订单状态在同一事务，避免"扣了库存但订单没更新"

### 第三层：失败恢复

"占坑成功，事务失败" 不能永远卡在 `processing`：

- 幂等记录增加 `status = processing | done | failed`
- `processing` 设置超时窗口（如 2 分钟）
- 超时后允许重试抢占，或转人工补偿队列

> 不要把 `processing` 永久当成"已成功"。

### 第四层：可观测性

| 指标 | 含义 |
|------|------|
| `idempotency_hit_total` | 命中重复次数，反映上游重试压力 |
| `idempotency_processing_timeout_total` | 处理超时次数，及时卡单发现 |
| `idempotency_failed_total` | 业务失败次数，评估补偿成本 |

## 总结

幂等是一次完整的工程设计，不是一行 `if order.Status == "paid"`：

1. 用**业务语义**定义幂等键
2. 用**唯一约束**做并发硬防线
3. 用**事务**保证关键状态原子变化
4. 用**失败恢复 + 监控**闭环兜底

如果你的系统现在只做了"状态判断"，建议下一步补上 **唯一约束 + 失败恢复**——这两件事决定了系统在重试风暴中的表现。

## 延伸阅读

- [Go 服务幂等设计实战（博客原文）](/blog/posts/2026/02/28/)
- [Concurrency and Idempotency Patterns (AWS)](https://aws.amazon.com/builders-library/making-retries-safe-with-idempotent-APIs/)

---

*维护人：yiiewang · 最后更新：2026-07-01*
