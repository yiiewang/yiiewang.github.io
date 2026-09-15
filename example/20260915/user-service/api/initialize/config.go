package initialize

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/api/global"
)

func GetEnvInfo(env string) bool {
	viper.AutomaticEnv()
	return viper.GetBool(env)
}

func Config() {
	b := GetEnvInfo("MS_DEBUG")
	configFilePrefix := "config"
	// 默认从 user-service 模块根目录运行（go run ./api），路径相对于运行时工作目录
	configFileName := fmt.Sprintf("%s/%s-prod.yaml", "api", configFilePrefix)
	if b {
		configFileName = fmt.Sprintf("%s/%s-debug.yaml", "api", configFilePrefix)
	}

	v := viper.New()
	v.SetConfigFile(configFileName)
	if err := v.ReadInConfig(); err != nil {
		zap.S().Panicf(err.Error())
	}
	if err := v.Unmarshal(global.ServerConfig); err != nil {
		zap.S().Panicf(err.Error())
	}

	go func() {
		v.WatchConfig()
		for {
			v.OnConfigChange(func(in fsnotify.Event) {
				zap.S().Infof("config file change: %s", in.Name)

				if err := v.ReadInConfig(); err != nil {
					zap.S().Panicf(err.Error())
				}
				if err := v.Unmarshal(global.ServerConfig); err != nil {
					zap.S().Panicf(err.Error())
				}
			})
		}
	}()
}
