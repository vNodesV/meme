# docker build . -t MeMeCosmos/meme:latest
# docker run --rm -it MeMeCosmos/meme:latest /bin/sh
FROM golang:1.22.10-alpine3.19 AS build-env
ARG TARGETARCH

# this comes from standard alpine nightly file
#  https://github.com/rust-lang/docker-rust-nightly/blob/master/alpine3.12/Dockerfile
# with some changes to support our toolchain, etc
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3 ca-certificates build-base git
RUN set -eux; apk add --no-cache $PACKAGES;

# NOTE: add these to run with LEDGER_ENABLED=true
# RUN apk add libusb-dev linux-headers

WORKDIR /go/src/github.com/MeMeCosmos/meme

COPY . .

# See https://github.com/CosmWasm/wasmvm/releases
ARG WASMVM_VERSION=v1.0.1
ARG WASMVM_SHA256_AARCH64=
ARG WASMVM_SHA256_X86_64=
ADD https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN if [ -n "$WASMVM_SHA256_AARCH64" ]; then echo "$WASMVM_SHA256_AARCH64  /lib/libwasmvm_muslc.aarch64.a" | sha256sum -c -; fi
RUN if [ -n "$WASMVM_SHA256_X86_64" ]; then echo "$WASMVM_SHA256_X86_64  /lib/libwasmvm_muslc.x86_64.a" | sha256sum -c -; fi

# Copy the library you want to the final location that will be found by the linker flag `-lwasmvm_muslc`
RUN case "${TARGETARCH}" in \
  amd64) arch="x86_64" ;; \
  arm64) arch="aarch64" ;; \
  *) arch="${TARGETARCH}" ;; \
  esac && \
  cp /lib/libwasmvm_muslc.${arch}.a /lib/libwasmvm_muslc.a

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc GOOS=linux GOARCH=${TARGETARCH} LEDGER_ENABLED=true make build

FROM alpine:3.19

RUN apk add --no-cache ca-certificates bash
RUN addgroup -S memed && adduser -S memed -G memed -u 10001
WORKDIR /home/memed

COPY --from=build-env /go/src/github.com/MeMeCosmos/meme/build/memed /usr/bin/memed

USER memed

# rest server
EXPOSE 1317 9090
# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657

CMD ["memed", "version"]

