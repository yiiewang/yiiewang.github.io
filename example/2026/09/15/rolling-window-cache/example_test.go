package rollingwindowcache_test

import (
	"fmt"
	"log"
	"time"

	rollingwindowcache "github.com/yiiewang/yiiewang.github.io/example/2026/09/15/rolling-window-cache"
)

// ExampleStandardRollingWindowCache_basic 基本使用：写入 + 两种查询
func ExampleStandardRollingWindowCache_basic() {
	// 容量 1000 个区块、16 个分片、时间窗口 1 小时
	cache := rollingwindowcache.NewStandardRollingWindowCache(1000, 16, time.Hour)

	block := &rollingwindowcache.BlockEntry{
		Height:       12345,
		Hash:         "0x1234567890abcdef",
		Timestamp:    time.Now().Unix(),
		Transactions: []string{"tx1", "tx2", "tx3"},
		TxCount:      3,
	}
	if err := cache.AddBlock(block); err != nil {
		log.Fatal(err)
	}

	if retrieved, found := cache.GetBlockByHeight(12345); found {
		fmt.Printf("Found block with hash: %s\n", retrieved.Hash)
	}
	if retrieved, found := cache.GetBlockByTxID("tx1"); found {
		fmt.Printf("Transaction tx1 is in block: %d\n", retrieved.Height)
	}

	// Output:
	// Found block with hash: 0x1234567890abcdef
	// Transaction tx1 is in block: 12345
}

// ExampleStandardRollingWindowCache_monitoring 监控指标：请求数 = 命中 + 未命中
func ExampleStandardRollingWindowCache_monitoring() {
	cache := rollingwindowcache.NewStandardRollingWindowCache(100, 8, time.Hour)

	for i := uint64(1); i <= 10; i++ {
		_ = cache.AddBlock(&rollingwindowcache.BlockEntry{
			Height:       i,
			Hash:         fmt.Sprintf("hash_%d", i),
			Timestamp:    time.Now().Unix(),
			Transactions: []string{fmt.Sprintf("tx_%d_1", i), fmt.Sprintf("tx_%d_2", i)},
			TxCount:      2,
		})
	}

	cache.GetBlockByHeight(1)  // 命中
	cache.GetBlockByHeight(5)  // 命中
	cache.GetBlockByHeight(99) // 未命中

	metrics := cache.GetMetrics()

	fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
	fmt.Printf("Cache hits: %d\n", metrics.CacheHits)
	fmt.Printf("Cache misses: %d\n", metrics.CacheMisses)
	fmt.Printf("Hit rate: %.1f%%\n", cache.GetHitRate()*100)
	fmt.Printf("Current size: %d\n", metrics.CurrentSize)

	// Output:
	// Total requests: 3
	// Cache hits: 2
	// Cache misses: 1
	// Hit rate: 66.7%
	// Current size: 10
}

// ExampleStandardRollingWindowCache_timeWindow 时间窗口查询
func ExampleStandardRollingWindowCache_timeWindow() {
	cache := rollingwindowcache.NewStandardRollingWindowCache(100, 4, time.Hour)

	now := time.Now().Unix()

	// 5 个区块，每个相差 1 分钟
	for i := uint64(1); i <= 5; i++ {
		_ = cache.AddBlock(&rollingwindowcache.BlockEntry{
			Height:       i,
			Hash:         fmt.Sprintf("hash_%d", i),
			Timestamp:    now - int64(i)*60,
			Transactions: []string{fmt.Sprintf("tx_%d", i)},
			TxCount:      1,
		})
	}

	// 最近 3 分钟
	blocks := cache.GetBlocksInTimeWindow(now-180, now)
	fmt.Printf("Blocks in last 3 minutes: %d\n", len(blocks))

	// 最近 3 个区块（最新在前）
	recent := cache.GetRecentBlocks(3)
	fmt.Printf("Recent 3 blocks count: %d\n", len(recent))
	fmt.Printf("Most recent block height: %d\n", recent[0].Height)

	// Output:
	// Blocks in last 3 minutes: 3
	// Recent 3 blocks count: 3
	// Most recent block height: 5
}

// ExampleStandardRollingWindowCache_eviction 淘汰机制：容量满后最旧区块出局
func ExampleStandardRollingWindowCache_eviction() {
	cache := rollingwindowcache.NewStandardRollingWindowCache(3, 2, time.Hour)

	for i := uint64(1); i <= 3; i++ {
		_ = cache.AddBlock(&rollingwindowcache.BlockEntry{
			Height:       i,
			Hash:         fmt.Sprintf("hash_%d", i),
			Timestamp:    time.Now().Unix(),
			Transactions: []string{fmt.Sprintf("tx_%d", i)},
			TxCount:      1,
		})
	}
	fmt.Printf("Cache size after adding 3 blocks: %d\n", cache.Size())

	_ = cache.AddBlock(&rollingwindowcache.BlockEntry{
		Height:       4,
		Hash:         "hash_4",
		Timestamp:    time.Now().Unix(),
		Transactions: []string{"tx_4"},
		TxCount:      1,
	})
	fmt.Printf("Cache size after adding 4th block: %d\n", cache.Size())

	if _, found := cache.GetBlockByHeight(1); !found {
		fmt.Println("Block 1 has been evicted")
	}
	if _, found := cache.GetBlockByHeight(4); found {
		fmt.Println("Block 4 exists in cache")
	}

	// Output:
	// Cache size after adding 3 blocks: 3
	// Cache size after adding 4th block: 3
	// Block 1 has been evicted
	// Block 4 exists in cache
}
