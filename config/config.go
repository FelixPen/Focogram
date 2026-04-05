package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name string
		Port string
	}
	Database struct {
		Dsn            string
		Max_open_conns int
		Max_idle_conns int
	}
	Redis struct {
		Addr     string
		Password string
		DB       int
	}
}

var AppConfig *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: %v, using environment variables", err)
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		log.Fatalf("配置文件解析错误: %v", err)
	}

	if envDsn := os.Getenv("DB_DSN"); envDsn != "" {
		AppConfig.Database.Dsn = envDsn
	}
	if envRedisAddr := os.Getenv("REDIS_ADDR"); envRedisAddr != "" {
		AppConfig.Redis.Addr = envRedisAddr
	}
	if envRedisPwd := os.Getenv("REDIS_PASSWORD"); envRedisPwd != "" {
		AppConfig.Redis.Password = envRedisPwd
	}

	initDB()

	initRedis()
}
