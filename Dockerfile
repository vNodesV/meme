# docker build . -t MeMeCosmos/meme:latest
# docker run --rm -it MeMeCosmos/meme:latest /bin/sh
FROM golang:1.22.10-alpine3.20 AS build-env
ARG TARGETARCH
ARG WASMVM_ARCH

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
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0-beta10/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0-beta10/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep 5b7abfdd307568f5339e2bea1523a6aa767cf57d6a8c72bc813476d790918e44
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep 2f44efa9c6c1cda138bd1f46d8d53c5ebfe1f4a53cf3457b01db86472c4917ac

# Copy the library you want to the final location that will be found by the linker flag `-lwasmvm_muslc`
RUN set -eux; \
  if [ -z "$WASMVM_ARCH" ]; then \
    case "$TARGETARCH" in \
      amd64) WASMVM_ARCH=x86_64 ;; \
      arm64) WASMVM_ARCH=aarch64 ;; \
      *) echo "unsupported TARGETARCH=$TARGETARCH" && exit 1 ;; \
    esac; \
  fi; \
  cp "/lib/libwasmvm_muslc.${WASMVM_ARCH}.a" /lib/libwasmvm_muslc.a

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc GOOS=linux GOARCH="$TARGETARCH" LEDGER_ENABLED=true make build

FROM alpine:3.20

RUN apk add --no-cache ca-certificates bash
RUN addgroup -S memed && adduser -S -G memed memed
WORKDIR /home/memed

COPY --from=build-env /go/src/github.com/MeMeCosmos/meme/build/memed /usr/bin/memed
RUN chown -R memed:memed /home/memed
USER memed


# rest server
EXPOSE 1317 9090
# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657

CMD ["memed", "version"]


