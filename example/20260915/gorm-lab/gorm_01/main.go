package main

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 演示用的默认 DSN（本地容器 MySQL）；真实项目请通过 DEMO_MYSQL_DSN 覆盖。
const defaultDSN = "root:root@tcp(127.0.0.1:3306)/db_test?charset=utf8mb4&parseTime=True&loc=Local"

type TBBank struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"column:user_name;type:varchar(20);index:idx_user_name"`
	Balance int    // 字段必须导出，未导出的字段 gorm 会静默忽略
}

// gorm_01：连接 + 自定义 logger（不建表，只验证能连上）
func main() {
	dsn := os.Getenv("DEMO_MYSQL_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Silent,
			Colorful:      true,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("连接 MySQL 失败（可用 DEMO_MYSQL_DSN 指定连接串）: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get sql.DB: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("ping 失败: %v", err)
	}
	log.Println("MySQL 连接成功")
}
