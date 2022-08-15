package main

import (
	"os"

	"github.com/daneshvar/go-logger"
	loggerinflux "github.com/daneshvar/go-logger-influx"
	"github.com/spf13/viper"
)

type LoggerConfig struct {
	Console *logger.ConsoleConfig
	Influx  *loggerinflux.Config
}

type config struct {
	Addr          string
	DisableHealth bool
	Cache         bool
	Services      map[string]string
	Ignores       []string
	Logger        LoggerConfig
}

func loadConfig(log *logger.Logger) *config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "configs/"
	}

	log.Infov("Loading", "path", path)

	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("reflection")

	if err := v.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	c := &config{}

	if err := v.Unmarshal(c); err != nil {
		log.Fatal(err)
	}

	log.Infof("%+v", c)

	return c
}
