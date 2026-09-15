package wal

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/wal"
)

// TestWAL 覆盖 tidwall/wal 的核心操作：写、读、截断、重开。
//
// 数据目录用 t.TempDir()：测试结束自动清理，不会把数据文件提交进仓库
// （原稿固定用相对目录 "mylog"，还引入了测试间的隐式执行顺序）。
func TestWAL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mylog")

	log, err := wal.Open(path, nil)
	require.NoError(t, err)

	// 写：index 从 1 开始，必须严格递增
	require.NoError(t, log.Write(1, []byte("first entry")))
	require.NoError(t, log.Write(2, []byte("second entry")))
	require.NoError(t, log.Write(3, []byte("third entry")))

	first, err := log.FirstIndex()
	require.NoError(t, err)
	last, err := log.LastIndex()
	require.NoError(t, err)
	t.Logf("index range: [%d, %d]", first, last)

	// 读：按 index 随机读
	data, err := log.Read(1)
	require.NoError(t, err)
	require.Equal(t, "first entry", string(data))

	data, err = log.Read(3)
	require.NoError(t, err)
	require.Equal(t, "third entry", string(data))

	// 截断：TruncateFront 丢弃 2 之前的，TruncateBack 丢弃 3 之后的
	require.NoError(t, log.TruncateFront(2))
	require.NoError(t, log.TruncateBack(3))

	_, err = log.Read(1)
	require.Error(t, err, "index 1 已被 TruncateFront 截断，读取应当失败")

	data, err = log.Read(2)
	require.NoError(t, err)
	require.Equal(t, "second entry", string(data))

	require.NoError(t, log.Close())

	// 关掉再打开：数据还在——这就是 WAL 的意义
	log, err = wal.Open(path, nil)
	require.NoError(t, err)
	defer log.Close()

	data, err = log.Read(2)
	require.NoError(t, err)
	require.Equal(t, "second entry", string(data))
}
