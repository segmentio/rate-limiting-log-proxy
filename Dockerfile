FROM golang:1.14-alpine as builder
RUN apk add --update curl ca-certificates make git gcc g++ python
# Enable go modules
ENV GO111MODULE=on
# enable go proxy for faster builds
ENV GOPROXY=https://proxy.golang.org
ENV GOPATH "/go"
COPY . /go/src/github.com/segmentio/rate-limiting-log-proxy
RUN apk add --update --no-cache go git musl-dev && rm -rf /var/cache/apk/* &&     cd /go/src/github.com/segmentio/rate-limiting-log-proxy &&     go get github.com/kardianos/govendor &&     /go/bin/govendor sync &&     go build -o /usr/local/bin/rate-limiting-log-proxy &&     apk del --purge go git musl-dev &&     rm -rf /go/src/github.com/segmentio/rate-limiting-log-proxy
WORKDIR $GOPATH/src/github.com/segmentio/rate-limiting-log-proxy
COPY . $GOPATH/src/github.com/segmentio/rate-limiting-log-proxy
# this is an auto-generated build command
# based upon the first argument of the entrypoint in the existing dockerfile.  
# This will work in most cases, but it is important to note
# that in some situations you may need to define a different build output with the -o flag
# This comment may be safely removed
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s -extldflags "-static"' -o /rate-limiting-log-proxy
FROM 528451384384.dkr.ecr.us-west-2.amazonaws.com/segment-scratch
COPY --from=builder rate-limiting-log-proxy rate-limiting-log-proxy
ENTRYPOINT ["rate-limiting-log-proxy"]
