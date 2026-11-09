package rollingwindowcache

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// BenchmarkAddBlock 写入：并发追加（会持续触发淘汰）
func BenchmarkAddBlock(b *testing.B) {
	cache := NewStandardRollingWindowCache(10000, 64, time.Hour)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		height := uint64(0)
		for pb.Next() {
			height++
			_ = cache.AddBlock(createTestBlock(height, 10))
		}
	})
}

// BenchmarkGetBlockByHeight 按高度查询：只碰全局读锁
func BenchmarkGetBlockByHeight(b *testing.B) {
	cache := NewStandardRollingWindowCache(10000, 64, time.Hour)

	for i := uint64(1); i <= 5000; i++ {
		_ = cache.AddBlock(createTestBlock(i, 10))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			height := uint64(rand.Intn(5000) + 1)
			cache.GetBlockByHeight(height)
		}
	})
}

// BenchmarkGetBlockByTxID 按交易ID查询：先走分片锁，再走全局读锁
func BenchmarkGetBlockByTxID(b *testing.B) {
	cache := NewStandardRollingWindowCache(10000, 64, time.Hour)

	for i := uint64(1); i <= 1000; i++ {
		_ = cache.AddBlock(createTestBlock(i, 10))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			height := rand.Intn(1000) + 1
			txIndex := rand.Intn(10)
			cache.GetBlockByTxID(fmt.Sprintf("tx_%d_%d", height, txIndex))
		}
	})
}

// BenchmarkMixedOperations 混合负载：10% 写、80% 读、10% 取最近区块
func BenchmarkMixedOperations(b *testing.B) {
	cache := NewStandardRollingWindowCache(10000, 64, time.Hour)

	for i := uint64(1); i <= 2000; i++ {
		_ = cache.AddBlock(createTestBlock(i, 5))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		height := uint64(2000)
		for pb.Next() {
			operation := rand.Intn(100)
			switch {
			case operation < 10:
				height++
				_ = cache.AddBlock(createTestBlock(height, 5))
			case operation < 60:
				cache.GetBlockByHeight(uint64(rand.Intn(int(height)) + 1))
			case operation < 90:
				queryHeight := rand.Intn(int(height)) + 1
				cache.GetBlockByTxID(fmt.Sprintf("tx_%d_%d", queryHeight, rand.Intn(5)))
			default:
				cache.GetRecentBlocks(10)
			}
		}
	})
}

// BenchmarkEviction 淘汰路径：容量 1000、每次写入都触发淘汰
func BenchmarkEviction(b *testing.B) {
	cache := NewStandardRollingWindowCache(1000, 16, time.Hour)

	for i := uint64(1); i <= 1000; i++ {
		_ = cache.AddBlock(createTestBlock(i, 10))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		height := uint64(1000)
		for pb.Next() {
			height++
			_ = cache.AddBlock(createTestBlock(height, 10))
		}
	})
}

// BenchmarkShardContention 分片数对交易ID查询的影响：
// 分片越多，锁竞争越小，但哈希本身的开销不变。
func BenchmarkShardContention(b *testing.B) {
	testCases := []struct {
		name       string
		shardCount int
	}{
		{"1_Shard", 1},
		{"4_Shards", 4},
		{"16_Shards", 16},
		{"64_Shards", 64},
		{"256_Shards", 256},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			cache := NewStandardRollingWindowCache(10000, tc.shardCount, time.Hour)

			for i := uint64(1); i <= 1000; i++ {
				_ = cache.AddBlock(createTestBlock(i, 10))
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					height := rand.Intn(1000) + 1
					cache.GetBlockByTxID(fmt.Sprintf("tx_%d_%d", height, rand.Intn(10)))
				}
			})
		})
	}
}

// BenchmarkCompareWithMap 与“裸 map + RWMutex”对比：
// 单锁 map 在并发读下已经不差，滑动窗口缓存的额外成本主要花在索引维护与指标上。
func BenchmarkCompareWithMap(b *testing.B) {
	type simpleCache struct {
		data map[uint64]*BlockEntry
		mu   sync.RWMutex
	}

	plain := &simpleCache{data: make(map[uint64]*BlockEntry)}
	rolling := NewStandardRollingWindowCache(10000, 16, time.Hour)

	for i := uint64(1); i <= 5000; i++ {
		block := createTestBlock(i, 10)

		plain.mu.Lock()
		plain.data[i] = block
		plain.mu.Unlock()

		_ = rolling.AddBlock(block)
	}

	b.Run("SimpleMap", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				height := uint64(rand.Intn(5000) + 1)
				plain.mu.RLock()
				_ = plain.data[height]
				plain.mu.RUnlock()
			}
		})
	})

	b.Run("RollingCache", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				height := uint64(rand.Intn(5000) + 1)
				rolling.GetBlockByHeight(height)
			}
		})
	})
}
