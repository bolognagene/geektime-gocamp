//go:build k8s

// 使用 k8s 这个编译标签
package config

var Config = config{
	DB: DBConfig{
		// 本地连接
		DSN: "root:root@tcp(10.104.111.102:3308)/webook",
	},
	Redis: RedisConfig{
		Addr: "10.108.49.83:6380",
	},
}
