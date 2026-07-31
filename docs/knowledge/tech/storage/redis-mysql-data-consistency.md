---
title: Redis/MySQL 数据一致性
domain: tech
created: 2026-07-04
updated: 2026-07-04
tags:
  - Redis
  - MySQL
  - 缓存
  - 数据一致性
  - 分布式系统
difficulty: intermediate
summary: "Redis缓存与MySQL数据库的数据一致性解决方案，包括双写、删除缓存、延迟双删等策略"
---

# Redis/MySQL 数据一致性

> Redis缓存与MySQL数据库的数据一致性解决方案，包括双写、删除缓存、延迟双删等策略

## 背景

在现代Web应用中，Redis作为缓存层与MySQL作为持久化存储层的组合非常常见。这种架构虽然能显著提升性能，但也带来了数据一致性的挑战。当数据在Redis和MySQL之间需要保持同步时，如何确保两个数据源的一致性成为关键问题。

数据不一致通常发生在以下场景：写操作成功更新了MySQL但Redis缓存更新失败、并发读写操作导致的竞态条件、缓存失效策略不当等。

## 核心内容

### 数据不一致的根本原因

!!! abstract "核心挑战"
    数据不一致主要源于四个关键因素：
    
    1. **写操作时序问题**：==先写缓存还是先写数据库==的决策
    2. **并发读写竞态**：多个线程同时操作相同数据产生的==竞态条件==
    3. **网络分区**：缓存与数据库之间的==网络故障==
    4. **缓存失效策略**：==过期时间设置不当==导致的同步问题

### 一致性解决方案

=== "双写策略（Write-Through）"
    **原理**：在写操作时，同时更新缓存和数据库
    
    **流程**：
    
    1. 开启事务
    2. 更新MySQL数据库
    3. 更新Redis缓存
    4. 提交事务
    
    **优点**：数据一致性高
    
    **缺点**：性能开销大，两个写操作都需要成功
    
    **适用场景**：金融交易、账户余额等强一致性要求场景

=== "删除缓存策略（Cache-Aside）"
    **原理**：写操作时删除缓存，读操作时重新加载
    
    **流程**：
    
    **写操作**：
    
    1. 删除Redis缓存
    2. 更新MySQL数据库
    
    **读操作**：
    
    1. 先查Redis缓存
    2. 缓存未命中则查MySQL
    3. 将结果写入Redis
    
    **优点**：实现简单，性能较好

    **缺点**：存在短暂的缓存不一致窗口
    
    **适用场景**：大多数Web应用场景

=== "延迟双删策略"
    **原理**：在删除缓存后，延迟一段时间再次删除，解决并发竞态
    
    **流程**：
    
    1. 删除Redis缓存
    2. 更新MySQL数据库
    3. 延迟一段时间（如500ms）
    4. 再次删除Redis缓存
    
    **优点**：有效解决并发竞态问题
    **缺点**：实现复杂，延迟时间需要合理设置
    
    **适用场景**：高并发写入场景

=== "异步消息队列"
    **原理**：通过消息队列异步同步数据
    
    **流程**：
    
    1. 更新MySQL数据库
    2. 发送消息到消息队列
    3. 消费者从队列读取消息并更新Redis
    
    **优点**：解耦，提高系统可用性
    **缺点**：最终一致性，存在延迟
    
    **适用场景**：大数据量、异步处理场景

### 一致性级别选择

!!! info "一致性级别对比"
    | 一致性级别 | 描述 | 适用场景 | 性能影响 |
    |-----------|------|----------|----------|
    | **强一致性** | 实时同步，数据完全一致 | 金融交易、账户余额 | ⭐⭐⭐⭐⭐ |
    | **最终一致性** | 短暂延迟后达到一致 | 商品库存、用户信息 | ⭐⭐⭐ |
    | **弱一致性** | 不保证实时一致 | 浏览量、点赞数 | ⭐ |
    
    **选择建议**：
    - ==强一致性==：对数据准确性要求极高的场景
    - ==最终一致性==：大多数业务场景的平衡选择
    - ==弱一致性==：对实时性要求不高的统计类数据

## 示例

### Go语言实现删除缓存策略

!!! example "代码实现"
    === "写操作实现"
        ```go title="write_operation.go" linenums="1"
        // Write操作：先删缓存，再更新数据库
        func UpdateUser(userID int, userData User) error {
            // 1. 删除缓存
            cacheKey := fmt.Sprintf("user:%d", userID)
            err := redisClient.Del(cacheKey).Err()
            if err != nil {
                log.Printf("删除缓存失败: %v", err)
                // 可以选择继续执行或返回错误
            }
            
            // 2. 更新数据库
            _, err = db.Exec("UPDATE users SET name = ?, email = ? WHERE id = ?", 
                userData.Name, userData.Email, userID)
            if err != nil {
                return fmt.Errorf("更新数据库失败: %v", err)
            }
            
            return nil
        }
        ```
    
    === "读操作实现"
        ```go title="read_operation.go" linenums="1" hl_lines="5-7 12-14"
        // Read操作：缓存未命中时加载
        func GetUser(userID int) (*User, error) {
            cacheKey := fmt.Sprintf("user:%d", userID)
            
            // 1. 先查缓存
            var user User
            err := redisClient.Get(cacheKey).Scan(&user)
            if err == nil {
                return &user, nil // 缓存命中
            }
            
            // 2. 缓存未命中，查数据库
            err = db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", userID).
                Scan(&user.ID, &user.Name, &user.Email)
            if err != nil {
                return nil, err
            }
            
            // 3. 写入缓存（设置过期时间）
            userJSON, _ := json.Marshal(user)
            redisClient.Set(cacheKey, userJSON, 30*time.Minute)
            
            return &user, nil
        }
        ```

## 注意事项

### ⚠️ 常见问题

!!! warning "缓存穿透"
    **问题**：查询不存在的数据导致频繁访问数据库
    
    **解决方案**：缓存空值或使用布隆过滤器

!!! warning "缓存雪崩"
    **问题**：大量缓存同时失效导致数据库压力骤增
    
    **解决方案**：设置不同的过期时间，使用热点数据永不过期

!!! warning "缓存击穿"
    **问题**：热点数据失效瞬间大量请求打到数据库
    
    **解决方案**：使用互斥锁或永不过期策略

### ✅ 最佳实践

!!! success "缓存策略优化"
    === "时间管理"
        **合理设置缓存过期时间**：根据业务特点设置TTL
        
        - 高频读数据：设置较长的TTL
        - 低频读数据：设置较短的TTL
        - 实时性要求高：设置短TTL或实时同步

    === "并发控制"
        **使用读写锁**：防止并发写操作导致的数据混乱
        
        - 读多写少：使用读写锁提高并发性能
        - 写操作频繁：考虑分布式锁

    === "监控告警"
        **监控缓存命中率**：及时发现缓存策略问题
        
        - 命中率低于80%：需要优化缓存策略
        - 命中率波动大：检查缓存失效策略

    === "容灾设计"
        **设计降级方案**：缓存故障时能正常服务
        
        - 直接访问数据库：保证基本功能可用
        - 限流保护：防止数据库被压垮

## 延伸阅读

- [Redis官方文档 - 缓存模式](https://redis.io/docs/manual/patterns/caching/)
- [《数据密集型应用系统设计》- 缓存一致性章节](https://dataintensive.net/)
- [MySQL与Redis数据一致性解决方案对比](https://medium.com/@david.truong/mysql-redis-consistency-7b7f5e5e5e5e)

---

*维护人：yiiewang · 最后更新：2026-07-04*