package main

import (
	"bytes"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func main() {
	initViperFromLocalConfigFile()
	//initViperFromReader()
	//initViperFromProgramArgument()
	//initViperFromRemote()
	server := InitWebServer()

	server.Run(":8077") // 监听并在 0.0.0.0:8077 上启动服务
}

func initViperFromLocalConfigFile() {
	// 第一种方式：直接设置ConfigFile
	// SetConfigFile explicitly defines the path, name and extension of the config file.
	// Viper will use this and not check any of the config paths.
	//viper.SetConfigFile("./config/dev.yaml")

	// 第二种方式：分别设置ConfigName, ConfigType, ConfigPath
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	// 实时监听配置变更
	viper.WatchConfig()
	// 只能告诉你文件变了，不能告诉你，文件的哪些内容变了
	viper.OnConfigChange(func(in fsnotify.Event) {
		// 比较好的设计，它会在 in 里面告诉你变更前的数据，和变更后的数据
		// 更好的设计是，它会直接告诉你差异。
		fmt.Println(in.Name, in.Op)
		fmt.Println(viper.GetString("TplId.code"))
		service.CodeTplId.Store(viper.GetString("TplId.code"))
	})
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("读取配置文件失败， %v", err))
	}
	keys := viper.AllKeys()
	println(keys)
	setting := viper.AllSettings()
	fmt.Println(setting)

	//otherViper := viper.New()
	//otherViper.SetConfigName("myjson")
	//otherViper.AddConfigPath("./config")
	//otherViper.SetConfigType("json")
}

func initViperFromReader() {
	viper.SetConfigType("yaml")
	cfg := `
db.mysql:
  dsn: "root:root@tcp(192.168.181.129:13316)/webook"

redis:
  addr: "192.168.181.129:6379"
`
	err := viper.ReadConfig(bytes.NewReader([]byte(cfg)))
	if err != nil {
		panic(err)
	}
	keys := viper.AllKeys()
	println(keys)
	setting := viper.AllSettings()
	fmt.Println(setting)
}

func initViperFromProgramArgument() {
	cfile := pflag.String("config", "config/config.yaml", "指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	keys := viper.AllKeys()
	println(keys)
	setting := viper.AllSettings()
	fmt.Println(setting)

}

func initViperFromRemote() {
	err := viper.AddRemoteProvider("etcd3",
		// 通过 webook 和其他使用 etcd 的区别出来
		"http://192.168.181.129:12379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
	keys := viper.AllKeys()
	println(keys)
	setting := viper.AllSettings()
	fmt.Println(setting)
}

/*func initUser(db *gorm.DB, cache redis_client.Cmdable) *web.UserHandler {
	userDao := dao.NewUserDAO(db)
	userCache := mycache.NewUserCache(cache, time.Minute*15)
	userRepo := repository.NewUserRepository(userDao, userCache)
	codeCache := mycache.NewCodeCache(cache)
	codeRepo := repository.NewCodeRepository(codeCache)
	memorySMSService := memory.NewService()
	codeService := service.NewCodeService(codeRepo, memorySMSService)
	userService := service.NewUserService(userRepo)
	userHandler := web.NewUserHandler(userService, codeService)
	return userHandler
}*/

/*func InitWebServer() *gin.Engine {
	server := gin.Default()

	return server
}*/
