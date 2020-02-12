FROM golang:stretch as builder

ADD . /go/src/github.com/smartcontractkit/substrate-adapter
RUN cd /go/src/github.com/smartcontractkit/substrate-adapter && go get && go build -o substrate-adapter

# Copy into a second stage container
FROM debian:stretch-slim

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
ENV PATH=$PATH:/root/.cargo/bin
RUN cargo install --force --git https://github.com/paritytech/substrate subkey

COPY --from=builder /go/src/github.com/smartcontractkit/substrate-adapter/substrate-adapter /usr/local/bin/

EXPOSE 8080
ENTRYPOINT ["substrate-adapter"]
