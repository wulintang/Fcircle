package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Server struct {
		Port      int    `mapstructure:"port"`
		SecretKey string `mapstructure:"secret_key"`
	} `mapstructure:"server"`

	Task struct {
		CronExpr string `mapstructure:"cron_expr"`
	} `mapstructure:"task"`

	RSS struct {
		ConfigURL  string `mapstructure:"config_url"`
		OutputFile string `mapstructure:"output_file"`
	} `mapstructure:"rss"`

	Log struct {
		File string `mapstructure:"file"`
	} `mapstructure:"log"`
}

var Config *AppConfig

func LoadConfig() *AppConfig {
	v := viper.New()

	//v.SetConfigName("config")
	//v.SetConfigType("yaml")
	//v.AddConfigPath("./config")
	//v.AddConfigPath("/app/config")
	//
	//if err := v.ReadInConfig(); err != nil {
	//	logrus.Warnf("配置文件读取失败：%v，将尝试从环境变量读取", err)
	//}

	v.AutomaticEnv()

	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.secret_key", "SECRET_KEY")
	v.BindEnv("task.cron_expr", "CRON_EXPR")
	v.BindEnv("rss.config_url", "RSS_CONFIG_URL")
	v.BindEnv("rss.output_file", "OUTPUT_FILE")
	v.BindEnv("log.file", "LOG_FILE")

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		logrus.Fatalf("配置解析失败：%v", err)
	}

	logrus.Infof("配置加载成功：端口 %d", cfg.Server.Port)

	Config = &cfg
	return Config
}
