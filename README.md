# rate-limiting-log-proxy

This proxy is designed to rate limit logs coming from docker in a per-container fashion.  It does this by acting like a syslog server, applying rate limiting, then forwarding logs to a local journald server.

The proxy tries to replicate as close as possible the log format coming out of the docker journald logging driver.

## Building

To build, just:

```bash
$ go build
```

## Tests

Run tests with

```bash
$ go test ./...
```

## Running

```bash
$ ./rate-limiting-log-proxy -help
Usage:
  rate-limiting-log-proxy [-h] [-help] [options...]

Options:
  -b int
    	Rate limit burst (default 500)

  -config-file source
    	Location to load the configuration file from.

  -d string
    	Docker host to connect to (default unix:///var/run/docker.sock)

  -i duration
    	Rate limit interval (default 5s)

  -p string
    	Port to host profiling endpoint (default "6060")

  -s string
    	Address to bind syslog server to (ex. udp://0.0.0.0:514) (default unixgram:///var/run/rate-limiting-log-proxy.sock)

```

## Configuring docker to send to proxy

To send to this proxy, you should configure your docker daemon with:

`--log-driver=syslog`
`--log-opts syslog-format=rfc3164`

Additionally you should set `--log-opts syslog-address` to the address you set on the proxy (default `unixgram:///var/run/rate-limiting-proxy.sock`)

If you would like to set a custom log tag, instead of using the normal `--log-opts tag=...` method, set a docker label during runtime like `docker run -l tag="{{ .ID }}"`.
