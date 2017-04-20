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

TODO: currently ports and destination of docker client are hardcoded, will pull this out into flags later

## Configuring docker to send to proxy

To send to this proxy, you should configure your docker daemon with:

`--log-driver=syslog`
`--log-opts syslog-format=rfc3164`
`--log-opts syslog-address=udp://localhost:10514`

If you would like to set a custom log tag, instead of using the normal `--log-opts tag=...` method, instead set a docker label during runtime like `docker run -l tag="{{ .ID }}"`.
