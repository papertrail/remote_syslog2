FROM golang:alpine AS builder

WORKDIR $GOPATH/src/github.com/papertrail/remote_syslog2

COPY ../.. .

RUN go get -v

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /remote_syslog2

FROM scratch

COPY --from=builder /remote_syslog2 /remote_syslog2
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT [ "/remote_syslog2", "-D", "--tls" ]
