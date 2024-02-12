package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	// 不需要Viper直接从程序里读
	// db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	// Viper读取配置的做法
	dsn := viper.GetString("db.mysql.dsn")
	if dsn == "" {
		dsn = "root:root@tcp(localhost:13316)/webook/default"
	}
	db, err := gorm.Open(mysql.Open(dsn))

	if err != nil {
		// 我只会在初始化过程中 panic
		// panic 相当于整个 goroutine 结束
		// 一旦初始化过程出错，应用就不要启动了
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
