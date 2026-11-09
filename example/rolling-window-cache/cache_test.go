package rollingwindowcache

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// createTestBlock 造一个测试区块：高度决定 hash 与交易ID前缀
func createTestBlock(height uint64, txCount int) *BlockEntry {
	transactions := make([]string, txCount)
	for i := 0; i < txCount; i++ {
		transactions[i] = fmt.Sprintf("tx_%d_%d", height, i)
	}

	return &BlockEntry{
		Height:       height,
		Hash:         fmt.Sprintf("hash_%d", height),
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		TxCount:      txCount,
	}
}

// TestNewStandardRollingWindowCache 测试缓存创建
func TestNewStandardRollingWindowCache(t *testing.T) {
	cache := NewStandardRollingWindowCache(100, 16, time.Hour)

	if cache.Capacity() != 100 {
		t.Errorf("Expected capacity 100, got %d", cache.Capacity())
	}
	if cache.Size() != 0 {
		t.Errorf("Expected size 0, got %d", cache.Size())
	}
	if len(cache.shards) != 16 {
		t.Errorf("Expected 16 shards, got %d", len(cache.shards))
	}
}

// TestAddBlock 测试添加区块
func TestAddBlock(t *testing.T) {
	cache := NewStandardRollingWindowCache(10, 4, time.Hour)

	block := createTestBlock(1, 3)
	if err := cache.AddBlock(block); err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}
	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}

	// 重复高度必须报错，否则索引会互相覆盖
	if err := cache.AddBlock(block); err == nil {
		t.Error("Expected error when adding duplicate block")
	}
}

// TestGetBlockByHeight 测试按高度获取区块
func TestGetBlockByHeight(t *testing.T) {
	cache := NewStandardRollingWindowCache(10, 4, time.Hour)

	block := createTestBlock(100, 2)
	if err := cache.AddBlock(block); err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}

	retrieved, found := cache.GetBlockByHeight(100)
	if !found {
		t.Fatal("Expected to find block")
	}
	if retrieved.Height != 100 {
		t.Errorf("Expected height 100, got %d", retrieved.Height)
	}

	if _, found = cache.GetBlockByHeight(999); found {
		t.Error("Expected not to find block")
	}
}

// TestGetBlockByTxID 测试按交易ID获取区块
func TestGetBlockByTxID(t *testing.T) {
	cache := NewStandardRollingWindowCache(10, 4, time.Hour)

	block := createTestBlock(200, 3)
	if err := cache.AddBlock(block); err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}

	retrieved, found := cache.GetBlockByTxID("tx_200_1")
	if !found {
		t.Fatal("Expected to find block by transaction ID")
	}
	if retrieved.Height != 200 {
		t.Errorf("Expected height 200, got %d", retrieved.Height)
	}

	if _, found = cache.GetBlockByTxID("nonexistent_tx"); found {
		t.Error("Expected not to find block")
	}
}

// TestRingBufferEviction 测试环形缓冲区淘汰机制
func TestRingBufferEviction(t *testing.T) {
	cache := NewStandardRollingWindowCache(3, 4, time.Hour)

	for i := uint64(1); i <= 3; i++ {
		if err := cache.AddBlock(createTestBlock(i, 2)); err != nil {
			t.Fatalf("Failed to add block %d: %v", i, err)
		}
	}
	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}

	// 第 4 个区块进场，第 1 个应该被淘汰
	if err := cache.AddBlock(createTestBlock(4, 2)); err != nil {
		t.Fatalf("Failed to add block 4: %v", err)
	}
	if cache.Size() != 3 {
		t.Errorf("Expected size 3 after eviction, got %d", cache.Size())
	}

	if _, found := cache.GetBlockByHeight(1); found {
		t.Error("Expected first block to be evicted")
	}
	if _, found := cache.GetBlockByHeight(4); !found {
		t.Error("Expected fourth block to exist")
	}

	// 被淘汰区块的交易索引也要一起清理，否则会留下指向空槽位的脏索引
	if _, found := cache.GetBlockByTxID("tx_1_0"); found {
		t.Error("Expected evicted block's transaction index to be cleaned")
	}
}

// TestGetRecentBlocks 测试获取最近区块
func TestGetRecentBlocks(t *testing.T) {
	cache := NewStandardRollingWindowCache(10, 4, time.Hour)

	for i := uint64(1); i <= 5; i++ {
		if err := cache.AddBlock(createTestBlock(i, 1)); err != nil {
			t.Fatalf("Failed to add block %d: %v", i, err)
		}
	}

	recent := cache.GetRecentBlocks(3)
	if len(recent) != 3 {
		t.Fatalf("Expected 3 recent blocks, got %d", len(recent))
	}
	// 最新在前
	if recent[0].Height != 5 {
		t.Errorf("Expected first block height 5, got %d", recent[0].Height)
	}
	if recent[2].Height != 3 {
		t.Errorf("Expected third block height 3, got %d", recent[2].Height)
	}

	// 请求数量超过缓存大小：返回全部，不报错
	if all := cache.GetRecentBlocks(10); len(all) != 5 {
		t.Errorf("Expected 5 blocks when requesting more than available, got %d", len(all))
	}
}

// TestGetBlocksInTimeWindow 测试时间窗口查询
func TestGetBlocksInTimeWindow(t *testing.T) {
	cache := NewStandardRollingWindowCache(10, 4, time.Hour)

	now := time.Now().Unix()

	// 每个区块相差 1 分钟
	for i := uint64(1); i <= 5; i++ {
		block := createTestBlock(i, 1)
		block.Timestamp = now - int64(i)*60
		if err := cache.AddBlock(block); err != nil {
			t.Fatalf("Failed to add block %d: %v", i, err)
		}
	}

	// 最近 3 分钟：区块 1(now-60)、2(now-120)、3(now-180)
	blocks := cache.GetBlocksInTimeWindow(now-180, now)
	if len(blocks) != 3 {
		t.Errorf("Expected 3 blocks in time window, got %d", len(blocks))
	}
}

// TestGetBlocksInWindow 测试使用构造参数 windowSize 的便捷查询
func TestGetBlocksInWindow(t *testing.T) {
	cache := NewStandardRollingWindowCache(10, 4, time.Minute)

	now := time.Now().Unix()

	// 两个在窗口内（30s / 50s 前），两个在窗口外（2min / 5min 前）
	for _, offset := range []int64{30, 50, 120, 300} {
		block := createTestBlock(uint64(offset), 1)
		block.Timestamp = now - offset
		if err := cache.AddBlock(block); err != nil {
			t.Fatalf("Failed to add block: %v", err)
		}
	}

	if blocks := cache.GetBlocksInWindow(); len(blocks) != 2 {
		t.Errorf("Expected 2 blocks in window, got %d", len(blocks))
	}
}

// TestConcurrentAccess 测试并发读写
func TestConcurrentAccess(t *testing.T) {
	cache := NewStandardRollingWindowCache(100, 16, time.Hour)

	const numGoroutines = 50
	const numOperations = 100

	var wg sync.WaitGroup

	// 并发写入
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				height := uint64(goroutineID*numOperations + j)
				_ = cache.AddBlock(createTestBlock(height, 3))
			}
		}(i)
	}

	// 并发读取
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				height := uint64(rand.Intn(numGoroutines * numOperations))
				cache.GetBlockByHeight(height)
				cache.GetBlockByTxID(fmt.Sprintf("tx_%d_0", height))
			}
		}()
	}

	wg.Wait()

	if cache.Size() > cache.Capacity() {
		t.Errorf("Cache size %d exceeds capacity %d", cache.Size(), cache.Capacity())
	}
}

// TestShardDistribution 测试分片分布
func TestShardDistribution(t *testing.T) {
	cache := NewStandardRollingWindowCache(100, 4, time.Hour)

	shardCounts := make(map[int]int)
	for _, txID := range []string{"tx1", "tx2", "tx3", "tx4", "tx5"} {
		shard := cache.getShard(txID)
		for i, s := range cache.shards {
			if s == shard {
				shardCounts[i]++
				break
			}
		}
	}

	if len(shardCounts) < 2 {
		t.Errorf("Expected at least 2 different shards, got %d", len(shardCounts))
	}
}

// TestMetrics 测试指标收集：一次请求只记一次
func TestMetrics(t *testing.T) {
	cache := NewStandardRollingWindowCache(10, 4, time.Hour)

	for i := uint64(1); i <= 3; i++ {
		if err := cache.AddBlock(createTestBlock(i, 2)); err != nil {
			t.Fatalf("Failed to add block %d: %v", i, err)
		}
	}

	cache.GetBlockByHeight(1)           // 命中
	cache.GetBlockByHeight(2)           // 命中
	cache.GetBlockByHeight(999)         // 未命中
	cache.GetBlockByTxID("tx_1_0")      // 命中（内部委托给高度查询，不应重复计数）
	cache.GetBlockByTxID("nonexistent") // 未命中

	metrics := cache.GetMetrics()

	if metrics.Insertions != 3 {
		t.Errorf("Expected 3 insertions, got %d", metrics.Insertions)
	}
	if metrics.TotalRequests != 5 {
		t.Errorf("Expected 5 total requests, got %d", metrics.TotalRequests)
	}
	if metrics.CacheHits != 3 {
		t.Errorf("Expected 3 cache hits, got %d", metrics.CacheHits)
	}
	if metrics.CacheMisses != 2 {
		t.Errorf("Expected 2 cache misses, got %d", metrics.CacheMisses)
	}
	if hitRate := cache.GetHitRate(); hitRate != 3.0/5.0 {
		t.Errorf("Expected hit rate %.2f, got %.2f", 3.0/5.0, hitRate)
	}
	if metrics.CurrentSize != 3 {
		t.Errorf("Expected current size 3, got %d", metrics.CurrentSize)
	}
	if metrics.MaxSize != 10 {
		t.Errorf("Expected max size 10, got %d", metrics.MaxSize)
	}
}

// TestClear 测试清空缓存
func TestClear(t *testing.T) {
	cache := NewStandardRollingWindowCache(10, 4, time.Hour)

	for i := uint64(1); i <= 5; i++ {
		if err := cache.AddBlock(createTestBlock(i, 2)); err != nil {
			t.Fatalf("Failed to add block %d: %v", i, err)
		}
	}
	if cache.Size() != 5 {
		t.Errorf("Expected size 5 before clear, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}
	if _, found := cache.GetBlockByHeight(1); found {
		t.Error("Expected not to find block after clear")
	}
	if _, found := cache.GetBlockByTxID("tx_1_0"); found {
		t.Error("Expected not to find transaction after clear")
	}
}

// TestEdgeCases 测试边界情况
func TestEdgeCases(t *testing.T) {
	cache := NewStandardRollingWindowCache(1, 1, time.Hour)

	// 空缓存
	if recent := cache.GetRecentBlocks(5); len(recent) != 0 {
		t.Errorf("Expected 0 recent blocks from empty cache, got %d", len(recent))
	}
	if blocks := cache.GetBlocksInTimeWindow(0, time.Now().Unix()); len(blocks) != 0 {
		t.Errorf("Expected 0 blocks in time window from empty cache, got %d", len(blocks))
	}

	// 容量为 1：每次写入都淘汰上一个
	if err := cache.AddBlock(createTestBlock(1, 1)); err != nil {
		t.Fatalf("Failed to add block 1: %v", err)
	}
	if err := cache.AddBlock(createTestBlock(2, 1)); err != nil {
		t.Fatalf("Failed to add block 2: %v", err)
	}

	if cache.Size() != 1 {
		t.Errorf("Expected size 1 in single-capacity cache, got %d", cache.Size())
	}
	if _, found := cache.GetBlockByHeight(1); found {
		t.Error("Expected first block to be evicted in single-capacity cache")
	}
	if _, found := cache.GetBlockByHeight(2); !found {
		t.Error("Expected second block to exist in single-capacity cache")
	}
}

// TestLargeTransactionBlock 测试大交易量区块：所有交易都能查到
func TestLargeTransactionBlock(t *testing.T) {
	cache := NewStandardRollingWindowCache(10, 8, time.Hour)

	if err := cache.AddBlock(createTestBlock(1, 1000)); err != nil {
		t.Fatalf("Failed to add large transaction block: %v", err)
	}

	for i := 0; i < 1000; i++ {
		txID := fmt.Sprintf("tx_1_%d", i)
		retrieved, found := cache.GetBlockByTxID(txID)
		if !found {
			t.Fatalf("Failed to find transaction %s", txID)
		}
		if retrieved.Height != 1 {
			t.Errorf("Wrong block height for transaction %s", txID)
		}
	}
}

// TestTimeWindowPrecision 测试时间窗口边界：闭区间
func TestTimeWindowPrecision(t *testing.T) {
	cache := NewStandardRollingWindowCache(10, 4, time.Hour)

	baseTime := int64(1640995200) // 2022-01-01 00:00:00 UTC

	for i := uint64(1); i <= 5; i++ {
		block := createTestBlock(i, 1)
		block.Timestamp = baseTime + int64(i)*10
		if err := cache.AddBlock(block); err != nil {
			t.Fatalf("Failed to add block %d: %v", i, err)
		}
	}

	// 窗口 [baseTime+15, baseTime+35] 内只有 baseTime+20、baseTime+30 两个
	blocks := cache.GetBlocksInTimeWindow(baseTime+15, baseTime+35)
	if len(blocks) != 2 {
		t.Errorf("Expected 2 blocks in precise time window, got %d", len(blocks))
	}

	for _, block := range blocks {
		if block.Timestamp < baseTime+15 || block.Timestamp > baseTime+35 {
			t.Errorf("Block timestamp %d is outside window", block.Timestamp)
		}
	}
}
