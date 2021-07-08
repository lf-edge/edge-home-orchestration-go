# Docker image for "edge-orchestration"
# TODO - need to reduce base image of edge-orchestration
### ubuntu:16.04 image size is 119MB
### alpine:3.6 image size is 4MB
ARG PLATFORM
FROM $PLATFORM/golang:1.16-alpine3.12 AS builder

# environment variables
ENV TARGET_DIR=/edge-orchestration

# set the working directory
WORKDIR $TARGET_DIR
ENV APP_BIN_DIR=bin
ENV APP_QEMU_DIR=$APP_BIN_DIR/qemu

#COPY $APP_QEMU_DIR/ /usr/bin/
COPY . .

# install required tools
RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --update --no-cache zeromq-dev libsodium-dev pkgconfig build-base git

RUN GO111MODULE=on go mod vendor
RUN make build_binary

FROM $PLATFORM/alpine:3.12

# environment variables
ENV TARGET_DIR=/edge-orchestration
ENV HTTP_PORT=56001
ENV MDNS_PORT=5353
ENV MNEDC_PORT=3334
ENV MNEDC_BROADCAST_PORT=3333
ENV ZEROCONF_PORT=42425
ENV APP_BIN_DIR=bin
ENV APP_NAME=edge-orchestration
ENV BUILD_DIR=build

# set the working directory
WORKDIR $TARGET_DIR

# copy files
COPY --from=builder $TARGET_DIR/$APP_BIN_DIR/$APP_NAME $TARGET_DIR/
COPY --from=builder $TARGET_DIR/$BUILD_DIR/package/run.sh $TARGET_DIR/
RUN mkdir -p $TARGET_DIR/res/

# install required tools
RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --update --no-cache zeromq-dev libsodium-dev pkgconfig build-base git net-tools iproute2

# expose ports
EXPOSE $HTTP_PORT $MDNS_PORT $ZEROCONF_PORT $MNEDC_PORT $MNEDC_BROADCAST_PORT

# kick off the edge-orchestration container
CMD ["sh", "run.sh"]
