package configs

import (
	"log"

	"github.com/spf13/viper"
)

type WechatConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
}
type OssConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	BucketName      string `mapstructure:"bucket_name"`
}
type Config struct {
	Wechat WechatConfig `mapstructure:"wechat"`
	Oss    OssConfig    `mapstructure:"oss"`
}

var Cfg *Config

func InitConfig() {
	viper.SetConfigName("config") // 文件名不带扩展名
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")        // 当前目录查找
	viper.AddConfigPath("./config") // 或 config 目录中查找

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}
}
