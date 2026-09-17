package leveldb

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
)

// TestLevelDB 基本读写删：数据目录用 t.TempDir()，不再依赖本机绝对路径。
func TestLevelDB(t *testing.T) {
	db, err := leveldb.OpenFile(filepath.Join(t.TempDir(), "myleveldb"), nil)
	require.NoError(t, err)
	defer db.Close()

	// 空库读不到：错误是 leveldb.ErrNotFound，不是 nil
	_, err = db.Get([]byte("key"), nil)
	require.ErrorIs(t, err, leveldb.ErrNotFound)

	require.NoError(t, db.Put([]byte("key"), []byte("value"), nil))

	data, err := db.Get([]byte("key"), nil)
	require.NoError(t, err)
	require.Equal(t, "value", string(data))

	require.NoError(t, db.Delete([]byte("key"), nil))

	_, err = db.Get([]byte("key"), nil)
	require.ErrorIs(t, err, leveldb.ErrNotFound)
}

// TestCompareLevelDB 逐 key 比对两个库：校验数据复制/迁移后的完整性。
func TestCompareLevelDB(t *testing.T) {
	src, err := leveldb.OpenFile(filepath.Join(t.TempDir(), "src"), nil)
	require.NoError(t, err)
	defer src.Close()

	dst, err := leveldb.OpenFile(filepath.Join(t.TempDir(), "dst"), nil)
	require.NoError(t, err)
	defer dst.Close()

	for i := 0; i < 100; i++ {
		key := []byte(fmt.Sprintf("key_%03d", i))
		val := []byte(fmt.Sprintf("value_%03d", i))
		require.NoError(t, src.Put(key, val, nil))
		require.NoError(t, dst.Put(key, val, nil))
	}

	ok, err := compare(src, dst)
	require.NoError(t, err)
	require.True(t, ok, "两个库应当完全一致")

	// 改一个 value，比对必须能发现差异
	require.NoError(t, dst.Put([]byte("key_050"), []byte("tampered"), nil))

	ok, err = compare(src, dst)
	require.NoError(t, err)
	require.False(t, ok, "存在差异时比对应当失败")

	// 少一个 key，同样要能发现
	require.NoError(t, dst.Delete([]byte("key_099"), nil))

	ok, err = compare(src, dst)
	require.NoError(t, err)
	require.False(t, ok, "key 数量不一致时比对应当失败")
}

// compare 比对两个库是否一致：key 集合相同，且每个 value 的 md5 相同。
// value 只存 md5，避免大 value 把内存吃满。
func compare(a, b *leveldb.DB) (bool, error) {
	keysA, err := dump(a)
	if err != nil {
		return false, err
	}
	keysB, err := dump(b)
	if err != nil {
		return false, err
	}

	if len(keysA) != len(keysB) {
		return false, nil
	}
	for k, va := range keysA {
		vb, ok := keysB[k]
		if !ok || !bytes.Equal(va, vb) {
			return false, nil
		}
	}
	return true, nil
}

// dump 把整个库读进 map[key]md5(value)
func dump(db *leveldb.DB) (map[string][]byte, error) {
	out := make(map[string][]byte)

	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		sum := md5.Sum(iter.Value())
		out[string(iter.Key())] = sum[:]
	}

	return out, iter.Error()
}
