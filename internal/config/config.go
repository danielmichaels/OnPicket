package config

import (
	"github.com/joeshaw/envdecode"
	"log"
	"time"
)

type Conf struct {
	Server  serverConf
	Db      dbConf
	Limiter limiter
	AppConf appConf
}

type limiter struct {
	Enabled bool          `env:"RATE_LIMIT_ENABLED,default=true"`
	Rps     int           `env:"RATE_LIMIT_RPS,default=10"`
	BackOff time.Duration `env:"RATE_LIMIT_BACKOFF,default=20s"`
}

type dbConf struct {
	DbCollection       string `env:"DATABASE_COLLECTION,default=domain"`
	DbName             string `env:"DATABASE_NAME,default=openrecce"`
	DbNameAuth         string `env:"DATABASE_NAME,default=admin"`
	DbUsername         string `env:"DATABASE_USERNAME,default=root"`
	DbPassword         string `env:"DATABASE_PASSWORD,default=root"`
	DbConnectionString string `env:"DATABASE_CONNECTION_STRING,default=mongodb.services"`
	DbPort             int    `env:"DATABASE_PORT,default=27017"`
}
type serverConf struct {
	Port         int           `env:"SERVER_PORT,default=9898"`
	TimeoutRead  time.Duration `env:"SERVER_TIMEOUT_READ,default=5s"`
	TimeoutWrite time.Duration `env:"SERVER_TIMEOUT_WRITE,default=10s"`
	TimeoutIdle  time.Duration `env:"SERVER_TIMEOUT_IDLE,default=15s"`
}
type appConf struct {
	LogLevel   string `env:"LOG_LEVEL,default=info"`
	LogConcise bool   `env:"LOG_CONCISE,default=true"`
	LogJson    bool   `env:"LOG_JSON,default=false"`
	LogCaller  bool   `env:"LOG_CALLER,default=false"`
}

// AppConfig Setup and install the applications' configuration environment variables
func AppConfig() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}
	return &c
}
