package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/gormx"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/prometheus"
	"time"
)

func InitDB(l logger.Logger) *gorm.DB {
	// 不需要Viper直接从程序里读
	// db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	// Viper读取配置的做法
	dsn := viper.GetString("db.mysql.dsn")
	if dsn == "" {
		dsn = "root:root@tcp(localhost:13316)/webook/default"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 缺了一个 writer
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			// 慢查询阈值，只有执行时间超过这个阈值，才会使用
			// 50ms， 100ms
			// SQL 查询必然要求命中索引，最好就是走一次磁盘 IO
			// 一次磁盘 IO 是不到 10ms
			SlowThreshold:             time.Millisecond * 10,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false, // 如果想在log里看到查询时传入的值，则设置为false
			LogLevel:                  glogger.Info,
		}),
	})

	if err != nil {
		// 我只会在初始化过程中 panic
		// panic 相当于整个 goroutine 结束
		// 一旦初始化过程出错，应用就不要启动了
		panic(err)
	}

	// Gorm的prometheus监控插件，用来监控数据库的一些参数（空闲连接啥的）
	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		StartServer:     false,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"Threads_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	// Gormx的prometheus监控SQL执行时间
	gcb := gormx.NewGormCallbacks(db, "Golang", "Webook", "webook")
	db.Use(gcb)

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}
