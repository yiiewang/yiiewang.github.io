package main

import (
	"errors"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 演示用的默认 DSN（本地容器 MySQL）；真实项目请通过 DEMO_MYSQL_DSN 覆盖。
const defaultDSN = "root:root@tcp(127.0.0.1:3306)/db_test?charset=utf8mb4&parseTime=True&loc=Local"

// TBBank 表模型：
//   - 表名默认是结构体名的蛇形复数（tb_banks），用 TableName 固定成 tb_bank
//   - Balance 必须导出，否则 gorm 静默忽略该列
type TBBank struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"column:user_name;type:varchar(20);index:idx_user_name"`
	Balance int
}

func (TBBank) TableName() string { return "tb_bank" }

// gorm_02：AutoMigrate + 一组 CRUD
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

	// 自动建表：字段、索引、列名都来自结构体 tag
	if err := db.AutoMigrate(&TBBank{}); err != nil {
		log.Fatalf("AutoMigrate 失败: %v", err)
	}
	log.Println("AutoMigrate 完成")

	// 清理上次运行的残留数据，保证多次运行结果一致
	if err := db.Where("user_name = ?", "cloaks").Delete(&TBBank{}).Error; err != nil {
		log.Fatalf("clean 失败: %v", err)
	}

	// Create
	bank := TBBank{Name: "cloaks", Balance: 100}
	if err := db.Create(&bank).Error; err != nil {
		log.Fatalf("create 失败: %v", err)
	}
	log.Printf("create: id=%d name=%s balance=%d", bank.ID, bank.Name, bank.Balance)

	// Read
	var got TBBank
	if err := db.Where("user_name = ?", "cloaks").First(&got).Error; err != nil {
		log.Fatalf("query 失败: %v", err)
	}
	log.Printf("query: %+v", got)

	// Update
	if err := db.Model(&got).Update("balance", 200).Error; err != nil {
		log.Fatalf("update 失败: %v", err)
	}

	// 再查一次确认更新生效
	var after TBBank
	if err := db.Where("user_name = ?", "cloaks").First(&after).Error; err != nil {
		log.Fatalf("query after update 失败: %v", err)
	}
	log.Printf("after update: balance=%d", after.Balance)

	// Delete
	if err := db.Delete(&after).Error; err != nil {
		log.Fatalf("delete 失败: %v", err)
	}

	// 确认删除后查不到（First 返回 ErrRecordNotFound）
	var gone TBBank
	err = db.Where("user_name = ?", "cloaks").First(&gone).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("期望 ErrRecordNotFound，实际: %v", err)
	}
	log.Println("delete 校验通过：记录已不存在")
}
