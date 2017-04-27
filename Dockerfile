FROM alpine:3.5

ENV GOPATH "/go"
COPY . /go/src/github.com/segmentio/rate-limiting-log-proxy
RUN apk add --update --no-cache go git musl-dev && rm -rf /var/cache/apk/* && \
    cd /go/src/github.com/segmentio/rate-limiting-log-proxy && \
    go get github.com/kardianos/govendor && \
    /go/bin/govendor sync && \
    go build -o /usr/local/bin/rate-limiting-log-proxy && \
    apk del --purge go git musl-dev && \
    rm -rf /go/src/github.com/segmentio/rate-limiting-log-proxy
ENTRYPOINT ["rate-limiting-log-proxy"]
