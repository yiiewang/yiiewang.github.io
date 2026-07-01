# 教程类范例：Go 求解回文十进制数

> **原始文件**：`docs/blog/posts/2025/01/15.md`
> **类型**：教程 + 算法题解
> **字数**：约 1500 字（短）
> **特色**：开门见山 + 代码驱动 + 反转函数教学

---

```markdown
---
date: 2025-01-15
authors:
  - cloaks
categories:
  - 札记
tags:
  - Go
  - 算法
  - 回文数
comments: true
description: 使用 Go 语言求解同时满足十进制、二进制、八进制表示都是回文数的最小值问题。
---

# 回文十进制数

!!! tip

    如果把某个数的各个数字按相反的顺序排列，得到的数和原来的数相同，则这个数就是"回文数"。譬如 123454321 就是一个回文数。

**问题**: 求用十进制、二进制、八进制表示都是回文数的所有数字中，大于十进制数 10 的最小值。

<!-- more -->

## 思路

因为是二进制的回文数，所以如果最低位是 0，那么相应地最高位也是 0。但是，以 0 开头肯定是不恰当的，由此可知最低位为 1。如果用二进制表示时最低位为 1，那这个数一定是奇数，因此只考虑奇数的情况就可以。接下来可以简单地编写程序，从 10 的下一个数字 11 开始，按顺序搜索。

## 示例代码

```go title="is_palindrome_test.go"
package golang

import (
    "strconv"
    "testing"
)

func reverse(s string) string {
    // 将字符串转换为rune切片（处理多字节字符）
    runes := []rune(s)
    // 双指针法翻转rune切片
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}

func isPalindrome(num int64) bool {
    binary := strconv.FormatInt(num, 2)
    if reverse(binary) != binary {
        return false
    }

    octal := strconv.FormatInt(num, 8)
    if reverse(octal) != octal {
        return false
    }

    decimal := strconv.FormatInt(num, 10)
    ...
```

## 解题关键

- 进制转换：`strconv.FormatInt(num, base)` 一行搞定
- 反转字符串：双指针法，O(n) 时间
- 优化空间：**奇数优先**（如思路所述）

## 答案

```bash
# 程序输出
585 = 1001001001 (binary) = 1111 (octal)
```

585 是大于 10 的最小数，同时满足十进制（585）、二进制（1001001001）、八进制（1111）都是回文数。
```

---

## 拆解要点

### 1. `!!! tip` 开篇
先给"前置知识"（什么是回文数），让读者**不用跳转**就能继续读。

### 2. 思路段把"为什么"讲透
不是上来就给代码，而是先**推理**：
> 因为二进制的回文数最低位是 1 → 所以是奇数 → 只考虑奇数

这一段"推理"比代码本身更有教学价值。

### 3. 极简章节
只有 4 个章节：思路 → 示例代码 → 解题关键 → 答案。**短文章就该短**。

### 4. 代码块用 `title`
```go title="is_palindrome_test.go"
```
读者一眼能看出文件名，复制即用。

### 5. 答案用 ` ```bash ` 代码块
不是用文字说"答案是 585"，而是给**可执行的 bash 输出**，让读者可以自己验证。
