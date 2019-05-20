package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/goforbroke1006/live-scores-main-hub/pkg"
	"github.com/goforbroke1006/live-scores-main-hub/pkg/config"
	"github.com/goforbroke1006/live-scores-main-hub/pkg/handler"
)

var (
	handleAddr = flag.String("handle-addr", "localhost:8080", "")
	configFile = flag.String("config-file", "./config.yaml", "")
)

func init() {
	flag.Parse()
}

const warmUpTimeout = 2500

func main() {
	cfg, err := config.LoadFromFile(*configFile)
	if nil != err {
		log.Fatal(err.Error())
	}

	logger := &log.Logger{}
	svc := pkg.NewMainHubService(logger)

	go func() {
		http.HandleFunc("/ws", handler.GetHandlerMiddleware(svc))
		if err := http.ListenAndServe(*handleAddr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(warmUpTimeout * time.Millisecond)

	done := make(chan bool)
	for name, url := range cfg.Providers {
		svc.RegisterProvider(name, url)
	}
	<-done
}
