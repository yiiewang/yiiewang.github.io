---
title: 享元模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 结构型模式
---

# 🪶 享元模式：共享让内存"变轻"

> 享元模式是一种结构型设计模式，它通过共享多个对象所共有的相同状态，让你能在有限的内存容量中载入更多对象。

## 从生活场景说起

想象你在开发《反恐精英》游戏，10 个玩家（5v5）各有服装：

- 🔴 恐怖分子：红色服装
- 🟢 反恐精英：绿色服装

如果每个玩家都独立存储服装数据，就要存 10 份。但同阵营穿的是**同一套衣服**！

**享元模式的解法** ：共享服装对象，只存 2 份。

## 核心概念

| 类型 | 说明 | 示例 |
|------|------|------|
| **内在状态** | 可共享、不变的数据 | 服装颜色、纹理 |
| **外在状态** | 不可共享、易变的数据 | 玩家位置、血量 |

## 完整实现

=== "dress.go: 享元接口"

    ```go
    package main

    type Dress interface {
        getColor() string
    }
    ```

=== "terroristDress.go: 具体享元"

    ```go
    package main

    type TerroristDress struct {
        color string
    }

    func (t *TerroristDress) getColor() string {
        return t.color
    }

    func newTerroristDress() *TerroristDress {
        return &TerroristDress{color: "red"}
    }
    ```

=== "dressFactory.go: 享元工厂"

    ```go
    package main

    const (
        TerroristDressType        = "tDress"
        CounterTerrroristDressType = "ctDress"
    )

    type DressFactory struct {
        dressMap map[string]Dress
    }

    var dressFactorySingleInstance = &DressFactory{
        dressMap: make(map[string]Dress),
    }

    func (d *DressFactory) getDressByType(dressType string) (Dress, error) {
        if d.dressMap[dressType] != nil {
            return d.dressMap[dressType], nil
        }
        if dressType == TerroristDressType {
            d.dressMap[dressType] = newTerroristDress()
            return d.dressMap[dressType], nil
        }
        return nil, fmt.Errorf("Wrong dress type")
    }
    ```

=== "player.go: 情景类"

    ```go
    package main

    type Player struct {
        dress      Dress
        playerType string
        lat, long  int
    }

    func newPlayer(playerType, dressType string) *Player {
        dress, _ := getDressFactorySingleInstance().getDressByType(dressType)
        return &Player{playerType: playerType, dress: dress}
    }
    ```

## 优缺点

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| 节省大量内存 | 代码复杂度增加 |
| 提升性能 | 需要计算外在状态 |

## 与其他模式的关系

- **组合模式**：可用享元共享叶节点
- **单例模式**：享元工厂通常是单例

---

!!! tip "一句话总结"
    **享元模式就是"合租"**：大家共用客厅（内在状态），各自有卧室（外在状态）。
