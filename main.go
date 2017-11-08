package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"
	"net/url"

	"github.com/segmentio/conf"
	"github.com/segmentio/rate-limiting-log-proxy/container"
	"github.com/segmentio/rate-limiting-log-proxy/logger"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

type config struct {
	DockerHost        string        `conf:"d" help:"Docker host to connect to"`
	RateLimitInterval time.Duration `conf:"i" help:"Rate limit interval"`
	RateLimitBurst    int           `conf:"b" help:"Rate limit burst"`
	ProfilingPort     string        `conf:"p" help:"Port to host profiling endpoint"`
	SyslogAddress     string        `conf:"s" help:"Address to bind syslog server to (ex. udp://0.0.0.0:514)"`
}

// DefaultConfig are the defaults for command line flags
var DefaultConfig = config{
	DockerHost:        "unix:///var/run/docker.sock",
	RateLimitInterval: 5 * time.Second,
	RateLimitBurst:    500,
	ProfilingPort:     "6060",
	SyslogAddress:     "unixgram:///var/run/rate-limiting-log-proxy.sock",
}

func main() {
	conf.Load(&DefaultConfig)

	go func() {
		profilingAddress := fmt.Sprintf("localhost:%s", DefaultConfig.ProfilingPort)
		log.Printf("Starting profiling server at %s...", profilingAddress)
		http.ListenAndServe(profilingAddress, nil)
	}()

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC3164)
	if err := setupListener(server, DefaultConfig.SyslogAddress); err != nil {
		log.Fatal(err)
	}

	containerLookupService, err := container.NewDockerLookup(DefaultConfig.DockerHost)
	if err != nil {
		log.Fatal(err)
	}

	loggerFactory := logger.NewLoggerFactory(logger.Journald)
	handler := NewRateLimitingHandler(DefaultConfig.RateLimitInterval, DefaultConfig.RateLimitBurst, containerLookupService, loggerFactory)
	server.SetHandler(handler)

	server.Boot()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigchan:
		log.Printf("Received %s, shutting down", sig)
		server.Kill()
		// Unix sockets not cleaned up automatically
		u, err := url.Parse(DefaultConfig.SyslogAddress)
		if err == nil && u.Scheme == "unixgram" {
			os.Remove(u.Path)
		}
	}

}

func setupListener(server *syslog.Server, address string) error {
	u, err := url.Parse(address)
	if err != nil {
		return err
	}

	log.Printf("Starting syslog server on %s...", address)
	switch u.Scheme {
	case "tcp":
		return server.ListenTCP(u.Host)
	case "udp":
		return server.ListenUDP(u.Host)
	case "unixgram":
		return server.ListenUnixgram(u.Path)
	default:
		return fmt.Errorf("Did not recognize syslog server scheme '%s'", u.Scheme)
	}
}
