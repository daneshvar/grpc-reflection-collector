package main

import (
	"os"

	"github.com/daneshvar/go-logger"
	"github.com/spf13/viper"
)

type config struct {
	Addr          string
	DisableHealth bool
	Services      map[string]string
	Ignores       []string
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
