package main

import (
	"fmt"
	"time"

	rollingwindowcache "github.com/yiiewang/yiiewang.github.io/example/2026/09/15/rolling-window-cache"
)

// 最小场景：容量 5 的缓存写入 8 个区块，
// 观察淘汰、两级索引查询与监控指标。
func main() {
	cache := rollingwindowcache.NewStandardRollingWindowCache(5, 4, time.Minute)

	for h := uint64(1); h <= 8; h++ {
		block := &rollingwindowcache.BlockEntry{
			Height:       h,
			Hash:         fmt.Sprintf("hash_%d", h),
			Timestamp:    time.Now().Unix(),
			Transactions: []string{fmt.Sprintf("tx_%d_a", h), fmt.Sprintf("tx_%d_b", h)},
			TxCount:      2,
		}
		if err := cache.AddBlock(block); err != nil {
			fmt.Println("add block:", err)
		}
	}

	// 容量 5：只剩高度 4..8
	_, ok1 := cache.GetBlockByHeight(1)
	block8, ok8 := cache.GetBlockByHeight(8)
	fmt.Printf("height 1 in cache: %v\n", ok1)
	fmt.Printf("height 8 in cache: %v (hash=%s)\n", ok8, block8.Hash)

	// 二级索引：用交易ID直接定位到区块
	if b, ok := cache.GetBlockByTxID("tx_7_b"); ok {
		fmt.Printf("tx_7_b found in block %d\n", b.Height)
	}

	// 被淘汰区块的交易索引同步失效
	if _, ok := cache.GetBlockByTxID("tx_1_a"); !ok {
		fmt.Println("tx_1_a not found (evicted)")
	}

	recent := cache.GetRecentBlocks(3)
	fmt.Print("recent heights:")
	for _, b := range recent {
		fmt.Printf(" %d", b.Height)
	}
	fmt.Println()

	metrics := cache.GetMetrics()
	fmt.Printf("size=%d/%d insertions=%d evictions=%d hitRate=%.1f%%\n",
		metrics.CurrentSize, metrics.MaxSize, metrics.Insertions, metrics.Evictions, cache.GetHitRate()*100)
}
