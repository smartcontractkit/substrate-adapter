FROM golang:stretch as go-builder

ADD . /go/src/github.com/smartcontractkit/substrate-adapter
RUN cd /go/src/github.com/smartcontractkit/substrate-adapter && go get && go build -o substrate-adapter

# build Rust binaries separately
FROM debian:stretch-slim as rust-builder
# install tools and dependencies
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get upgrade -y && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        libssl1.1 \
        ca-certificates \
        curl && \
# apt cleanup
    apt-get autoremove -y && \
    apt-get clean && \
    find /var/lib/apt/lists/ -type f -not -name lock -delete;

RUN curl https://getsubstrate.io -sSf | bash -s -- --fast
RUN /root/.cargo/bin/cargo install --force --git https://github.com/paritytech/substrate subkey

FROM debian:stretch-slim

COPY --from=go-builder /go/src/github.com/smartcontractkit/substrate-adapter/substrate-adapter /usr/local/bin/
COPY --from=rust-builder /root/.cargo/bin /usr/local/bin/
COPY --from=rust-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 8080
ENTRYPOINT ["substrate-adapter"]
