FROM golang:stretch as go-builder

ADD . /go/src/github.com/smartcontractkit/substrate-adapter
RUN cd /go/src/github.com/smartcontractkit/substrate-adapter && go get && go build -o substrate-adapter

FROM debian:stretch-slim as ssl-certificates
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get upgrade -y && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        libssl1.1 \
        ca-certificates && \
# apt cleanup
    apt-get autoremove -y && \
    apt-get clean && \
    find /var/lib/apt/lists/ -type f -not -name lock -delete;

FROM parity/subkey:2.0.0 as subkey

FROM debian:stretch-slim

COPY --from=go-builder /go/src/github.com/smartcontractkit/substrate-adapter/substrate-adapter /usr/local/bin/
COPY --from=subkey /usr/local/bin/ /usr/local/bin/
COPY --from=ssl-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 8080
ENTRYPOINT ["substrate-adapter"]
