package mysql

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	// 驱动必须显式导入（原稿漏了这行，sql.Open("mysql") 必然报 unknown driver）
	_ "github.com/go-sql-driver/mysql"
)

// 连接串通过环境变量指定，默认指向本地容器 MySQL：
//
//	docker run -d --name demo-mysql \
//	  -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=db_test \
//	  -p 3306:3306 mysql:8
//
// 连不上时测试 skip，不阻塞本地/CI 的其他用例。
const defaultDSN = "root:root@tcp(127.0.0.1:3306)/db_test?parseTime=true"

func openDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("DEMO_MYSQL_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}

	db.SetMaxOpenConns(4)
	db.SetConnMaxLifetime(time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("MySQL 不可达，跳过该用例（可设置 DEMO_MYSQL_DSN）: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

// TestMySQLBinaryColumn 二进制列的读写：
// 写进去是 []byte，查出来只打 hex 是看不懂的，hex + ASCII 一起看才能确认内容。
func TestMySQLBinaryColumn(t *testing.T) {
	db := openDB(t)

	const ddl = `CREATE TABLE IF NOT EXISTS key_values (
		object_key   VARBINARY(64)  NOT NULL,
		object_value VARBINARY(512) NOT NULL,
		PRIMARY KEY (object_key)
	)`
	if _, err := db.Exec(ddl); err != nil {
		t.Fatalf("create table: %v", err)
	}

	key := []byte("im0")
	// 一段 protobuf 风格的二进制数据
	value := []byte("\x12\x1400000000000000000001\x18\t \xe7\x01")

	if _, err := db.Exec(
		"REPLACE INTO key_values (object_key, object_value) VALUES (?, ?)",
		key, value,
	); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var (
		gotKey   []byte
		gotValue []byte
	)
	row := db.QueryRow(
		"SELECT object_key, object_value FROM key_values WHERE object_key = ?", key,
	)
	if err := row.Scan(&gotKey, &gotValue); err != nil {
		if err == sql.ErrNoRows {
			t.Fatal("未查到数据")
		}
		t.Fatalf("scan: %v", err)
	}

	t.Logf("key:   hex=%s ascii=%q", hex.EncodeToString(gotKey), gotKey)
	t.Logf("value: hex=%s ascii=%q", hex.EncodeToString(gotValue), gotValue)

	if !bytes.Equal(gotValue, value) {
		t.Fatalf("value 不一致: got %x, want %x", gotValue, value)
	}
}

// TestSplitStatements 纯函数用例：把 SQL 脚本按分号切成语句。
//
// 原稿的正则少了 (?m) 多行标志，`;(?:\s*)$` 只会在整个脚本的末尾切一刀，
// 多条语句会被当成一条执行，必然报语法错误。
func TestSplitStatements(t *testing.T) {
	script := `
-- 建表
CREATE TABLE t (id INT);
INSERT INTO t VALUES (1);
INSERT INTO t VALUES
  (2),
  (3);
`

	stmts := splitStatements(script)

	if len(stmts) != 3 {
		t.Fatalf("期望切出 3 条语句，实际 %d 条: %#v", len(stmts), stmts)
	}
	if !strings.HasPrefix(stmts[0], "CREATE TABLE") {
		t.Errorf("第 1 条应当是建表语句: %q", stmts[0])
	}
	if !strings.Contains(stmts[2], "(2),") {
		t.Errorf("多行语句应当完整保留: %q", stmts[2])
	}
}

// splitStatements 去掉行注释后，按行尾的分号切分。
func splitStatements(script string) []string {
	reComment := regexp.MustCompile(`(?m)--[^\n]*$`)
	cleaned := reComment.ReplaceAllString(script, "")

	reSplit := regexp.MustCompile(`(?m);[ \t]*(?:\n|$)`)

	var out []string
	for _, part := range reSplit.Split(cleaned, -1) {
		if s := strings.TrimSpace(part); s != "" {
			out = append(out, s)
		}
	}
	return out
}
