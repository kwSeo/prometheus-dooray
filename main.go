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

func main() {
	var port int
	var bindAddr string
	var basePath string
	var incomingURL string
	flag.IntVar(&port, "server.port", 8080, "listen port of this server")
	flag.StringVar(&bindAddr, "server.bind-addr", "", "bind address")
	flag.StringVar(&basePath, "server.base-path", "", "base URL path of this server")
	flag.StringVar(&incomingURL, "dooray.incoming-url", "", "incoming URL of Dooray Messanger")
	flag.Parse()

	listenAddr := fmt.Sprintf("%s:%d", bindAddr, port)
	level.Info(logger).Log("msg", "Starting server.", "addr", listenAddr)

	http.Handle(basePath+"/api/v1/alerts", dooray.NewAlertmanagerHandler(incomingURL, logger))
	http.Handle(basePath+"/webhook", dooray.NewWebhookHandler(incomingURL, logger))
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	level.Info(logger).Log("msg", "Stopped.")
}
