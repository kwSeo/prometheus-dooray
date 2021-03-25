package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/kwseo/prometheus-dooray/pkg/dooray"
)

var logger = log.With(log.NewLogfmtLogger(os.Stdout), "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

type Config struct {
	Port     int
	BindAddr string
	BasePath string
	Dooray   dooray.Config

	set *flag.FlagSet
}

func NewConfig(set *flag.FlagSet) *Config {
	return &Config{
		set: set,
	}
}

func (c *Config) Load(args []string) error {
	c.set.IntVar(&c.Port, "server.port", 8080, "listen port of this server")
	c.set.StringVar(&c.BindAddr, "server.bind-addr", "", "bind address")
	c.set.StringVar(&c.BasePath, "server.base-path", "", "base URL path of this server")
	c.Dooray.RegisterFlags(c.set)
	return c.set.Parse(args)
}

func main() {
	cfg := NewConfig(flag.CommandLine)
	if err := cfg.Load(os.Args[1:]); err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	listenAddr := fmt.Sprintf("%s:%d", cfg.BindAddr, cfg.Port)
	level.Info(logger).Log("msg", "Starting server.", "addr", listenAddr)

	http.Handle(cfg.BasePath+"/api/v1/alerts", dooray.NewAlertmanagerHandler(cfg.Dooray, logger))
	http.Handle(cfg.BasePath+"/webhook", dooray.NewWebhookHandler(cfg.Dooray, logger))
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	level.Info(logger).Log("msg", "Stopped.")
}
