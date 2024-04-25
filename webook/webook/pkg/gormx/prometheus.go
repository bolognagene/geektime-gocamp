package gormx

import (
	promsdk "github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
	"time"
)

// GormCallbacks 为GORM框架加上Prometheus监控, 监控SQL语句执行时间
type GormCallbacks struct {
	vector *promsdk.SummaryVec
}

func NewGormCallbacks(gdb *gorm.DB, namespace string, subsystem string, db string) *GormCallbacks {
	vector := promsdk.NewSummaryVec(promsdk.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "gorm_query_duration",
		Help:      "统计 GORM 的执行时间",
		ConstLabels: map[string]string{
			"db": db,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	},
		// 如果是 JOIN 查询，table 就是 JOIN 在一起的
		// 或者 table 就是主表，A JOIN B，记录的是 A
		[]string{"type", "table"})

	gcb := &GormCallbacks{
		//db:     gdb,
		vector: vector,
	}
	promsdk.MustRegister(vector)

	return gcb
}

func (gcb *GormCallbacks) before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (gcb *GormCallbacks) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		startTime, ok := val.(time.Time)
		if !ok {
			return
		}

		table := db.Statement.Table
		gcb.vector.WithLabelValues(typ, table).
			Observe(float64(time.Duration(time.Since(startTime).Milliseconds())))
	}
}

func (gcb *GormCallbacks) registerAll(db *gorm.DB) {
	// 作用于 INSERT 语句
	err := db.Callback().Create().Before("*").Register("prometheus_create_before",
		gcb.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Create().After("*").Register("prometheus_create_after",
		gcb.after("create"))
	if err != nil {
		panic(err)
	}

	// 作用于 Update 语句
	err = db.Callback().Update().Before("*").Register("prometheus_update_before",
		gcb.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Update().After("*").Register("prometheus_update_after",
		gcb.after("update"))
	if err != nil {
		panic(err)
	}

	// 作用于 DELETE 语句
	err = db.Callback().Delete().Before("*").Register("prometheus_delete_before",
		gcb.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Delete().After("*").Register("prometheus_delete_after",
		gcb.after("delete"))
	if err != nil {
		panic(err)
	}

	// 作用于 Raw 语句
	err = db.Callback().Raw().Before("*").Register("prometheus_raw_before",
		gcb.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Raw().After("*").Register("prometheus_raw_after",
		gcb.after("raw"))
	if err != nil {
		panic(err)
	}

	// 作用于 Row 语句
	err = db.Callback().Row().Before("*").Register("prometheus_row_before",
		gcb.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Row().After("*").Register("prometheus_row_after",
		gcb.after("row"))
	if err != nil {
		panic(err)
	}
}

func (gcb *GormCallbacks) Initialize(db *gorm.DB) error {
	gcb.registerAll(db)
	return nil
}

func (gcb *GormCallbacks) Name() string {
	return "gormx_prometheus"
}
