package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/segmentio/rate-limiting-log-proxy/container"
	"github.com/segmentio/rate-limiting-log-proxy/logger"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

func main() {
	// Start http server for pprof
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Start syslog server
	server := syslog.NewServer()
	server.SetFormat(syslog.RFC3164)
	server.ListenUDP("0.0.0.0:10514")

	containerLookupService, err := container.NewDockerLookup("unix:///var/run/docker.sock")
	if err != nil {
		log.Fatal(err)
	}

	loggerFactory := logger.NewLoggerFactory(logger.Journald)
	handler := NewRateLimitingHandler(5*time.Second, 300, containerLookupService, loggerFactory)
	server.SetHandler(handler)

	server.Boot()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigchan:
		server.Kill()
	}

}
